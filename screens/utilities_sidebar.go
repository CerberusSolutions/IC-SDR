package screens

import (
	"fmt"
	"path/filepath"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"go-zero/simpleui"
)

const utilitiesRight = float32(334)

// UtilitiesSidebar keeps the three operational tools visible while a decoder
// remains selected in the lower workspace.
type UtilitiesSidebar struct {
	screen    *MainScreen
	controls  []simpleui.Element
	scan      *simpleui.Button
	scanLayer *simpleui.Switch
	scanRange *simpleui.Button
	scanMode  *simpleui.Button
	scanMem   *simpleui.Button
	groupPick *simpleui.Dropdown
	previous  *simpleui.Button
	next      *simpleui.Button
	recall    *simpleui.Button
	add       *simpleui.Button
	edit      *simpleui.Button
	remove    *simpleui.Button
	record    *simpleui.Button
	pause     *simpleui.Button
	skip      *simpleui.Button
	folder    *simpleui.Button
}

func NewUtilitiesSidebar(screen *MainScreen) *UtilitiesSidebar {
	p := &UtilitiesSidebar{screen: screen}
	button := func(id, label string, x, y, w, h float32, action func()) *simpleui.Button {
		b := simpleui.NewButton(id, x, y, w, h, label, 11)
		b.SetColors(colors.panelAlt, colors.border, colors.text)
		b.OnClick(action)
		p.controls = append(p.controls, b)
		return b
	}
	p.scan = button("utilityScan", "INICIAR", 40, 339, 132, 34, screen.scanPanel.ToggleRunning)
	p.scanLayer = simpleui.NewSwitch("utilityScanLayer", 211, 237, 103, 28, "CAPA FFT", true, 11)
	p.scanLayer.SetTrackColors(colors.panelAlt, colors.green)
	p.scanLayer.OnChange(func(active bool) { screen.scanPanel.overlayVisible = active })
	p.controls = append(p.controls, p.scanLayer)
	p.scanRange = button("utilityScanRange", "RANGO FFT", 180, 339, 134, 34, func() {
		x := screen.centerFrequencyHz
		half := screen.spanHz / 2
		screen.scanPanel.minimumHz, screen.scanPanel.maximumHz = x-half+screen.spanHz/10, x+half-screen.spanHz/10
		screen.markSettingsDirty()
	})
	p.scanMode = button("utilityScanMode", "REANUDAR", 40, 298, 132, 31, func() {
		values := []string{"AUTO", "DELAY", "HOLD"}
		i := slices.Index(values, screen.scanPanel.resume)
		screen.scanPanel.resume = values[(i+1)%len(values)]
		screen.markSettingsDirty()
	})
	p.scanMem = button("utilityScanMemory", "AJUSTE MEM", 180, 298, 134, 31, func() {
		screen.scanPanel.centerToMemory = !screen.scanPanel.centerToMemory
		screen.markSettingsDirty()
	})
	p.add = button("utilityMemoryAdd", "+ MEMORIA", 178, 405, 136, 31, screen.memoryPanel.openSaveModal)
	addGroup := button("utilityMemoryAddGroup", "+ GRUPO", 40, 405, 130, 31, screen.memoryPanel.openNewGroupModal)
	_ = addGroup
	p.groupPick = simpleui.NewDropdown("utilityMemoryFilter", 40, 442, 274, 32, "TODAS", screen.memoryPanel.groups, 13)
	p.groupPick.OnChange(func(_ int, group string) {
		screen.memoryPanel.selectedGroup, screen.memoryPanel.selected, screen.memoryPanel.scrollOffset = group, -1, 0
	})
	p.controls = append(p.controls, p.groupPick)
	p.previous = button("utilityMemoryPrevious", "◄", 40, 700, 42, 31, func() { p.moveMemory(-1) })
	p.next = button("utilityMemoryNext", "►", 88, 700, 42, 31, func() { p.moveMemory(1) })
	p.recall = button("utilityMemoryRecall", "SINTONIZAR", 178, 700, 136, 31, screen.memoryPanel.tuneSelected)
	p.edit = button("utilityMemoryEdit", "EDITAR", 40, 737, 130, 31, screen.memoryPanel.editSelection)
	p.remove = button("utilityMemoryDelete", "ELIMINAR", 178, 737, 136, 31, screen.memoryPanel.openDeleteModal)
	p.remove.SetColors(actionClearFill, colors.red, colors.text)
	p.record = button("utilityRecord", "GRABAR", 40, 836, 82, 32, screen.recorderPanel.ToggleRecording)
	p.record.SetColors(actionStopFill, colors.red, colors.text)
	p.pause = button("utilityPause", "PAUSA", 128, 836, 76, 32, screen.recorderPanel.TogglePause)
	p.skip = button("utilitySkip", "SQL", 210, 836, 54, 32, screen.recorderPanel.ToggleSkipSilence)
	p.folder = button("utilityFolder", "DIR", 270, 836, 44, 32, screen.recorderPanel.OpenFolder)
	return p
}

func (p *UtilitiesSidebar) moveMemory(delta int) {
	indices := p.screen.memoryPanel.filteredIndices()
	p.screen.memoryPanel.moveSelection(indices, delta)
}

func (p *UtilitiesSidebar) Draw() {
	drawPanel(24, 215, utilitiesRight-24, 685)
	p.drawSection(225, 158, "ESCÁNER", colors.cyan)
	p.drawSection(393, 385, "MEMORIAS", colors.blue)
	p.drawSection(788, 102, "GRABADOR", colors.red)

	scan := p.screen.scanPanel
	if scan.running {
		p.scan.SetLabel("DETENER")
		p.scan.SetColors(actionStopFill, colors.red, colors.text)
	} else {
		p.scan.SetLabel("INICIAR")
		p.scan.SetColors(actionStartFill, colors.green, colors.text)
	}
	simpleui.DrawTextStyled(scan.displayStatus(), 40, 258, 12, simpleui.FontSemiBold, func() rl.Color {
		if scan.running {
			return colors.green
		}
		return colors.muted
	}())
	simpleui.DrawText(fmt.Sprintf("%.3f–%.3f MHz · SQL %d", float64(scan.minimumHz)/1e6, float64(scan.maximumHz)/1e6, p.screen.squelchThreshold), 40, 278, 11, colors.muted)
	p.scanMode.SetLabel("REANUDAR " + scan.resume)
	if scan.centerToMemory {
		p.scanMem.SetLabel("MEM ON")
	} else {
		p.scanMem.SetLabel("MEM OFF")
	}

	memory := p.screen.memoryPanel
	if !slices.Equal(p.groupPick.Items(), memory.groups) {
		p.groupPick.SetItems(memory.groups)
	}
	for i, group := range memory.groups {
		if group == memory.selectedGroup && p.groupPick.SelectedIndex() != i {
			p.groupPick.SetSelected(i)
		}
	}
	p.drawMemoryTable(memory)

	state := p.screen.recorder.State()
	status, statusColor := "PREPARADO", colors.muted
	if state.Recording {
		status, statusColor = "GRABANDO", colors.red
		if state.Paused {
			status, statusColor = "PAUSADO", colors.orange
		}
		p.record.SetLabel("DETENER")
	} else {
		p.record.SetLabel("GRABAR")
	}
	p.pause.SetEnabled(state.Recording)
	if state.Paused {
		p.pause.SetLabel("CONTINUAR")
	} else {
		p.pause.SetLabel("PAUSA")
	}
	if state.SkipSquelchSilence {
		p.skip.SetLabel("SQL ON")
	} else {
		p.skip.SetLabel("SQL OFF")
	}
	rl.DrawCircle(44, 817, 5, statusColor)
	simpleui.DrawTextStyled(status+"  "+formatRecordingDuration(state.DurationSeconds), 56, 809, 12, simpleui.FontMono, colors.text)
	if len(state.RecentFiles) > 0 {
		simpleui.DrawText(trimMemory(filepath.Base(state.RecentFiles[0]), 27), 40, 873, 10, colors.muted)
	}
}

func (p *UtilitiesSidebar) drawMemoryTable(memory *MemoryPanel) {
	x, y, w, rowH := float32(40), float32(481), float32(274), float32(29)
	rl.DrawRectangleLinesEx(rl.Rectangle{X: x, Y: y, Width: w, Height: 213}, 1, colors.border)
	simpleui.DrawTextStyled("GRUPO", x+7, y+8, 13, simpleui.FontSemiBold, colors.cyan)
	simpleui.DrawTextStyled("NOMBRE", x+78, y+8, 13, simpleui.FontSemiBold, colors.cyan)
	simpleui.DrawTextStyled("MHz", x+181, y+8, 13, simpleui.FontSemiBold, colors.cyan)
	indices := memory.filteredIndices()
	memory.scrollOffset = min(max(memory.scrollOffset, 0), max(len(indices)-6, 0))
	end := min(memory.scrollOffset+6, len(indices))
	for row, index := range indices[memory.scrollOffset:end] {
		m := memory.memories[index]
		yy := y + 32 + float32(row)*rowH
		if index == memory.selected {
			rl.DrawRectangleRec(rl.Rectangle{X: x + 1, Y: yy - 4, Width: w - 2, Height: rowH}, rl.Color{R: 18, G: 72, B: 105, A: 255})
		}
		simpleui.DrawText(trimMemory(memoryGroup(m), 9), x+7, yy, 13, memory.groupColor(memoryGroup(m)))
		simpleui.DrawText(trimMemory(m.Name, 11), x+78, yy, 13, colors.text)
		simpleui.DrawText(fmt.Sprintf("%.5f", float64(m.FrequencyHz)/1e6), x+181, yy, 13, colors.text)
	}
}

func (p *UtilitiesSidebar) UpdateInput() {
	m := p.screen.memoryPanel
	mouse := simpleui.MousePosition()
	indices := m.filteredIndices()
	if mouse.X >= 40 && mouse.X <= 314 && mouse.Y >= 513 && mouse.Y < 687 {
		if wheel := rl.GetMouseWheelMove(); wheel != 0 {
			m.scrollOffset -= int(wheel)
			m.scrollOffset = min(max(m.scrollOffset, 0), max(len(indices)-6, 0))
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
			row := int((mouse.Y - 513) / 29)
			pos := m.scrollOffset + row
			if pos >= 0 && pos < len(indices) {
				index := indices[pos]
				now := rl.GetTime()
				m.selected = index
				if m.lastClicked == index && now-m.lastClickAt <= .42 {
					m.tuneSelected()
					m.lastClicked = -1
				} else {
					m.lastClicked, m.lastClickAt = index, now
				}
			}
		}
	}
}

func (p *UtilitiesSidebar) drawSection(y, height float32, title string, accent rl.Color) {
	rl.DrawRectangleRoundedLinesEx(rl.Rectangle{X: 31, Y: y, Width: utilitiesRight - 38, Height: height}, .05, 6, 1, colors.border)
	simpleui.DrawTextStyled(title, 40, y+8, 12, simpleui.FontSemiBold, accent)
}
