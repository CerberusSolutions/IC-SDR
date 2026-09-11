package screens

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"go-zero/internal/aprs"
	"go-zero/internal/resources"
	"go-zero/simpleui"

	"go-zero/internal/i18n"
)

type APRSPanel struct {
	screen           *MainScreen
	controls         []simpleui.Element
	view             string
	buttons          map[string]*simpleui.Button
	scroll, selected int
	feedback         string
	feedbackUntil    time.Time
	snapshotPath     string
	lastSnapshot     string
	viewer           *exec.Cmd
	viewerDone       chan struct{}
}

// APRS view identifiers. These are stored in settings.json, so they are
// language-neutral keys rather than the Spanish words the buttons show; the
// button labels are translated for display like any other interface string.
const (
	aprsViewPackets  = "PACKETS"
	aprsViewStations = "STATIONS"
	aprsViewMessages = "MESSAGES"
	aprsViewRadar    = "RADAR"
	aprsViewRaw      = "RAW"
)

// aprsViews lists every valid identifier in the order the buttons appear.
var aprsViews = []string{aprsViewPackets, aprsViewStations, aprsViewMessages, aprsViewRadar, aprsViewRaw}

// legacyAPRSViews maps the Spanish identifiers written by IC-SDR up to
// settings version 1 onto the current ones.
var legacyAPRSViews = map[string]string{
	"PAQUETES":   aprsViewPackets,
	"ESTACIONES": aprsViewStations,
	"MENSAJES":   aprsViewMessages,
}

func NewAPRSPanel(screen *MainScreen) *APRSPanel {
	p := &APRSPanel{screen: screen, view: aprsViewPackets, buttons: map[string]*simpleui.Button{}, selected: -1, snapshotPath: resources.WritablePath("cache", "aprs-captures.json")}
	items := []struct {
		id, label string
		w         float32
		color     rl.Color
	}{{aprsViewPackets, "PAQUETES", 120, colors.blue}, {aprsViewStations, "ESTACIONES", 130, colors.panelAlt}, {aprsViewMessages, "MENSAJES", 120, colors.panelAlt}, {aprsViewRadar, "RADAR", 100, colors.panelAlt}, {aprsViewRaw, "RAW", 85, colors.panelAlt}}
	x := float32(40)
	for _, item := range items {
		b := simpleui.NewButton("aprs"+item.id, x, 660, item.w, 40, item.label, uiControlFontSize)
		b.SetColors(item.color, colors.border, colors.text)
		view := item.id
		b.OnClick(func() {
			p.view = view
			p.screen.aprsView = view
			p.scroll = 0
			p.selected = -1
			p.style()
			p.screen.markSettingsDirty()
		})
		p.buttons[item.id] = b
		p.controls = append(p.controls, b)
		x += item.w + 10
	}
	europe := simpleui.NewButton("aprsEurope", 650, 660, 185, 40, "APRS EU 144.800", uiControlFontSize)
	europe.SetColors(rl.Color{R: 68, G: 49, B: 92, A: 255}, rl.Color{R: 175, G: 145, B: 245, A: 255}, colors.text)
	europe.OnClick(func() { p.tune(144_800_000) })
	popout := simpleui.NewButton("aprsPopout", 845, 660, 170, 40, "ABRIR TABLA", uiControlFontSize)
	popout.SetColors(colors.blue, colors.border, colors.text)
	popout.OnClick(p.openViewer)
	export := simpleui.NewButton("aprsExport", 1025, 660, 180, 40, "EXPORTAR CSV", uiControlFontSize)
	export.SetColors(colors.green, colors.border, colors.background)
	export.OnClick(p.export)
	clearButton := simpleui.NewButton("aprsClear", 1215, 660, 125, 40, "LIMPIAR", uiControlFontSize)
	clearButton.SetColors(actionClearFill, colors.red, colors.text)
	clearButton.OnClick(func() {
		if screen.receiver != nil {
			screen.receiver.ClearAPRSPackets()
		}
		p.scroll, p.selected = 0, -1
		p.say("HISTORIAL LIMPIADO")
	})
	p.controls = append(p.controls, europe, popout, export, clearButton)
	p.SetVisible(false)
	return p
}
func (p *APRSPanel) tune(hz int64) {
	p.screen.frequencyHz, p.screen.centerFrequencyHz, p.screen.centerMode = hz, hz, true
	p.screen.bandCategory, p.screen.bandName = "HAM", "2 m"
	if p.screen.band != nil {
		p.screen.band.SetLabel("BAND  2 m")
	}
	if p.screen.vfoModeSwitch != nil {
		p.screen.vfoModeSwitch.SetActive(false)
	}
	if p.screen.receiver != nil {
		p.screen.receiver.SetCenterFrequency(hz)
		p.screen.receiver.SetDemodulator("NFM", hz, p.screen.demodBandwidthHz)
		p.screen.receiver.ConfigureAPRS(true, hz, p.screen.demodBandwidthHz)
	}
	p.screen.waterfall.Reset()
	p.screen.markSettingsDirty()
	p.say("SINTONIZADO EN APRS EUROPA")
}
func (p *APRSPanel) style() {
	for id, b := range p.buttons {
		fill := colors.panelAlt
		if id == p.view {
			fill = rl.Color{R: 20, G: 100, B: 98, A: 255}
		}
		b.SetColors(fill, colors.cyan, colors.text)
	}
}
func (p *APRSPanel) SetVisible(v bool) {
	for _, c := range p.controls {
		c.SetVisible(v)
	}
	if v {
		p.style()
	}
}
func (p *APRSPanel) Enter() {
	p.screen.selectMode("NFM")
	if p.screen.receiver != nil {
		p.screen.receiver.ConfigureAPRS(true, p.screen.frequencyHz, p.screen.demodBandwidthHz)
	}
}
func (p *APRSPanel) Leave() {
	if p.screen.receiver != nil {
		p.screen.receiver.ConfigureAPRS(false, 0, 0)
	}
}
func (p *APRSPanel) Close() {
	p.Leave()
	if p.viewer != nil && p.viewer.Process != nil {
		_ = p.viewer.Process.Kill()
	}
}
func (p *APRSPanel) packets() []aprs.Packet {
	if p.screen.receiver == nil {
		return nil
	}
	return p.screen.receiver.APRSPackets()
}
func (p *APRSPanel) Tick() {
	if p.viewerDone != nil {
		select {
		case <-p.viewerDone:
			p.viewer, p.viewerDone = nil, nil
		default:
		}
	}
	if p.screen.activeTool != "APRS" || p.screen.viewMode != 1 || p.screen.overlayOpen() {
		return
	}
	p.writeSnapshot(p.packets())
	if p.screen.receiver != nil {
		p.screen.receiver.ConfigureAPRS(true, p.screen.frequencyHz, p.screen.demodBandwidthHz)
	}
	packets := p.filtered()
	visible := 5
	if p.view == aprsViewRadar {
		return
	}
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		p.scroll = min(max(p.scroll-int(wheel), 0), max(len(packets)-visible, 0))
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		p.selected = min(p.selected+1, len(packets)-1)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		p.selected = max(p.selected-1, 0)
	}
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		m := simpleui.MousePosition()
		for row := 0; row < visible && p.scroll+row < len(packets); row++ {
			if rl.CheckCollisionPointRec(m, rl.Rectangle{X: 280, Y: 724 + float32(row)*17, Width: 900, Height: 17}) {
				p.selected = p.scroll + row
				simpleui.PlayActivationFeedback()
			}
		}
	}
}
func (p *APRSPanel) writeSnapshot(packets []aprs.Packet) {
	data, err := json.Marshal(packets)
	if err != nil {
		p.lastSnapshot = ""
		p.say("ERROR AL PREPARAR LA TABLA")
		return
	}
	value := string(data)
	if value == p.lastSnapshot {
		return
	}
	p.lastSnapshot = value
	if replaceLiveSnapshot(p.snapshotPath, data) != nil {
		p.lastSnapshot = ""
	}
}
func (p *APRSPanel) openViewer() {
	p.writeSnapshot(p.packets())
	if p.viewer != nil && p.viewer.Process != nil {
		focusRTL433Viewer(p.viewer.Process.Pid)
		p.say("TABLA YA ABIERTA")
		return
	}
	executable, err := os.Executable()
	if err != nil {
		p.say("ERROR AL ABRIR TABLA")
		return
	}
	cmd := exec.Command(executable, "--aprs-viewer", p.snapshotPath)
	cmd.SysProcAttr = rtl433ViewerProcessAttributes()
	if err = cmd.Start(); err != nil {
		p.say("ERROR AL ABRIR TABLA")
		return
	}
	p.viewer = cmd
	p.viewerDone = make(chan struct{})
	done := p.viewerDone
	go func() { _ = cmd.Wait(); close(done) }()
	p.say("TABLA ABIERTA")
}
func (p *APRSPanel) filtered() []aprs.Packet {
	packets := p.packets()
	if p.view == aprsViewMessages {
		out := packets[:0]
		for _, packet := range packets {
			if packet.Type == "MESSAGE" || packet.Type == "ACK" || packet.Type == "REJ" {
				out = append(out, packet)
			}
		}
		return out
	}
	if p.view == aprsViewStations {
		seen := map[string]bool{}
		out := []aprs.Packet{}
		for _, packet := range packets {
			if !seen[packet.Source] {
				seen[packet.Source] = true
				out = append(out, packet)
			}
		}
		return out
	}
	return packets
}
func (p *APRSPanel) DrawPanel() {
	status := aprs.Status{State: "NO DISPONIBLE", AudioLevel: -1}
	if p.screen.receiver != nil {
		status = p.screen.receiver.APRSStatus()
	}
	simpleui.DrawTextStyled("APRS RX · AFSK 1200 / AX.25", 40, 638, 16, simpleui.FontSemiBold, rl.Color{R: 55, G: 215, B: 195, A: 255})
	stateColor := colors.orange
	if status.State == "RECIBIENDO" {
		stateColor = colors.green
	} else if status.State == "ERROR" {
		stateColor = colors.red
	}
	rl.DrawCircle(720, 646, 6, stateColor)
	simpleui.DrawTextStyled(i18n.Tf("%s · %.3f MHz · NIVEL %s · RX %d · ERR %+.0f Hz", i18n.T(status.State), float64(p.screen.frequencyHz)/1e6, levelText(status.AudioLevel), status.PacketCount, status.FrequencyErrorHz), 735, 638, 12, simpleui.FontSemiBold, colors.muted)
	drawPanel(40, 712, 220, 94)
	simpleui.DrawTextStyled("RECEPTOR", 52, 721, 12, simpleui.FontSemiBold, colors.cyan)
	simpleui.DrawText(fmt.Sprintf("KISS  %s", i18n.T(map[bool]string{true: "CONECTADO", false: "ESPERANDO"}[status.KISS])), 52, 745, 13, colors.text)
	simpleui.DrawText(i18n.Tf("COLA %d/16 · DROP %d", status.Queued, status.Dropped), 52, 766, 12, colors.text)
	simpleui.DrawText(short(i18n.T(status.Detail), 28), 52, 787, 12, colors.muted)
	if p.view == aprsViewRadar {
		p.drawRadar()
		return
	}
	p.drawTable(p.filtered())
	if time.Now().Before(p.feedbackUntil) {
		simpleui.DrawTextStyled(p.feedback, 1360, 672, 12, simpleui.FontSemiBold, colors.green)
	}
}
func levelText(v int) string {
	if v < 0 {
		return "—"
	}
	return strconv.Itoa(v)
}
func (p *APRSPanel) drawTable(packets []aprs.Packet) {
	drawPanel(275, 712, 1275, 94)
	headers := []struct {
		x float32
		s string
	}{{288, "HORA"}, {365, "INDICATIVO"}, {485, "TIPO"}, {590, "DESTINO"}, {700, "RUTA"}, {900, "POSICIÓN / MENSAJE"}, {1230, "NIVEL"}}
	for _, h := range headers {
		simpleui.DrawTextStyled(h.s, h.x, 718, 12, simpleui.FontSemiBold, colors.cyan)
	}
	if len(packets) == 0 {
		simpleui.DrawText("ESPERANDO TRAMAS APRS", 730, 760, 14, colors.muted)
		return
	}
	p.scroll = min(p.scroll, max(0, len(packets)-5))
	for row := 0; row < 5 && p.scroll+row < len(packets); row++ {
		index := p.scroll + row
		packet := packets[index]
		y := 744 + float32(row)*13
		if index == p.selected {
			rl.DrawRectangle(282, int32(y-2), 1258, 15, rl.Color{R: 25, G: 83, B: 116, A: 220})
		}
		typeColor := rl.Color{R: 190, G: 145, B: 235, A: 255}
		if packet.Type == "POSITION" {
			typeColor = colors.green
		} else if packet.Type == "WEATHER" {
			typeColor = rl.Color{R: 75, G: 145, B: 255, A: 255}
		} else if packet.Type == "MESSAGE" {
			typeColor = colors.orange
		}
		simpleui.DrawTextStyled(packet.Received.Format("15:04:05"), 288, y, 12, simpleui.FontMono, colors.text)
		simpleui.DrawTextStyledRaw(short(packet.Source, 13), 365, y, 12, simpleui.FontSemiBold, colors.cyan)
		simpleui.DrawTextRaw(packet.Type, 485, y, 12, typeColor)
		simpleui.DrawTextRaw(short(packet.Destination, 12), 590, y, 12, colors.text)
		path := packet.Path
		if path == "" {
			path = "DIRECTO"
		}
		simpleui.DrawTextRaw(short(path, 25), 700, y, 12, colors.text)
		info := packet.Summary
		if p.view == aprsViewRaw {
			info = packet.Raw
		} else if p.view == aprsViewStations && packet.Coordinates != "—" {
			info = packet.Coordinates + " · " + packet.Summary
		}
		simpleui.DrawTextRaw(short(info, 43), 900, y, 12, colors.text)
		simpleui.DrawText(levelText(packet.ReceiveLevel), 1230, y, 12, colors.text)
	}
	if p.selected >= 0 && p.selected < len(packets) {
		packet := packets[p.selected]
		simpleui.DrawTextRaw(short(packet.Raw, 42), 1290, 744, 12, colors.muted)
	}
}
func (p *APRSPanel) drawRadar() {
	drawPanel(275, 712, 1275, 94)
	packets := p.filteredPositions()
	if len(packets) == 0 {
		simpleui.DrawText("RADAR OFFLINE · ESPERANDO POSICIONES", 680, 758, 14, colors.muted)
		return
	}
	minLat, maxLat, minLon, maxLon := packets[0].Latitude, packets[0].Latitude, packets[0].Longitude, packets[0].Longitude
	for _, v := range packets {
		minLat = min(minLat, v.Latitude)
		maxLat = max(maxLat, v.Latitude)
		minLon = min(minLon, v.Longitude)
		maxLon = max(maxLon, v.Longitude)
	}
	cx, cy := float32(910), float32(760)
	rl.DrawCircleLines(int32(cx), int32(cy), 37, colors.grid)
	rl.DrawCircleLines(int32(cx), int32(cy), 18, colors.grid)
	rl.DrawLine(int32(cx-45), int32(cy), int32(cx+45), int32(cy), colors.grid)
	rl.DrawLine(int32(cx), int32(cy-42), int32(cx), int32(cy+42), colors.grid)
	latSpan := max(maxLat-minLat, .01)
	lonSpan := max(maxLon-minLon, .01)
	for _, v := range packets {
		x := cx + float32((v.Longitude-(minLon+maxLon)/2)/lonSpan)*80
		y := cy - float32((v.Latitude-(minLat+maxLat)/2)/latSpan)*72
		rl.DrawCircle(int32(x), int32(y), 4, colors.green)
		simpleui.DrawTextRaw(v.Source, x+7, y-6, 12, colors.text)
	}
	simpleui.DrawText("VISTA RELATIVA OFFLINE", 290, 725, 12, colors.cyan)
	simpleui.DrawText(i18n.Tf("%d estaciones con posición", len(packets)), 290, 748, 13, colors.text)
	simpleui.DrawText("Sin conexión a Internet", 290, 772, 12, colors.muted)
}
func (p *APRSPanel) filteredPositions() []aprs.Packet {
	seen := map[string]bool{}
	out := []aprs.Packet{}
	for _, v := range p.packets() {
		if !seen[v.Source] && v.Coordinates != "—" {
			seen[v.Source] = true
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	return out
}
func (p *APRSPanel) say(s string) { p.feedback = s; p.feedbackUntil = time.Now().Add(3 * time.Second) }
func (p *APRSPanel) export() {
	path, err := exportAPRSCSV(p.filtered())
	if err != nil {
		p.say("ERROR AL EXPORTAR")
	} else {
		p.say("CSV: " + filepath.Base(path))
	}
}
func exportAPRSCSV(packets []aprs.Packet) (string, error) {
	dir := resources.WritablePath("exports", "aprs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "aprs-"+time.Now().Format("20060102-150405")+".csv")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, _ = f.Write([]byte{0xef, 0xbb, 0xbf})
	w := csv.NewWriter(f)
	w.Comma = ';'
	_ = w.Write(strings.Split("fecha;hora;indicativo;destino;ruta;tipo;simbolo;coordenadas;locator;rumbo;velocidad;altitud;destino_mensaje;id_mensaje;temperatura;humedad;presion;viento;lluvia;nivel;informacion;trama_raw", ";"))
	for _, p := range packets {
		_ = w.Write([]string{p.Received.Format("2006-01-02"), p.Received.Format("15:04:05"), p.Source, p.Destination, p.Path, p.Type, p.Symbol, p.Coordinates, p.Locator, p.Course, p.Speed, p.Altitude, p.MessageTarget, p.MessageID, p.Temperature, p.Humidity, p.Pressure, p.Wind, p.Rain, strconv.Itoa(p.ReceiveLevel), p.Summary, p.Raw})
	}
	w.Flush()
	return path, w.Error()
}
