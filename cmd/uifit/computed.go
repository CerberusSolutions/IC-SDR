package main

import "go-zero/simpleui"

// computedGeometry lists the places whose width is worked out at run time
// rather than written in the source, so the parser cannot see it. Each entry
// records the room the drawing code actually leaves for the text.
//
// Keep this in step with the panels it names: if one of them is re-laid out,
// the numbers here need the same change.
func computedGeometry() []measurement {
	var found []measurement
	add := func(where, text string, available float32, size int32, style simpleui.FontStyle) {
		found = append(found, measurement{where, "computed", text, available, size, style})
	}

	// band_selector.go categoryBounds: 280 then 260 wide, centred at size 17.
	for index, label := range []string{"RADIOAFICIONADO / HAM", "COMERCIALES", "ISM / LIBRE"} {
		width := float32(260)
		if index == 0 {
			width = 280
		}
		add("band_selector categories", label, width-12, 17, simpleui.FontRegular)
	}

	// filter_selector.go presetBounds: 164 wide, description centred at 11.
	for _, label := range []string{"ANCHO", "MEDIO", "ESTRECHO", "PERSONALIZADO", "AMPLIO", "COMPLETO", "REDUCIDO", "DMR 12,5", "TETRA 25"} {
		add("filter_selector presets", label, 164-10, 11, simpleui.FontRegular)
	}

	// tetra_viewer.go tabs: 179 wide at size 14, but the viewer runs with a
	// text scale of 1.55, so measure the size the glyphs are really drawn at.
	for _, label := range []string{"RED", "CELDAS", "GRUPOS", "USUARIOS", "MENSAJES", "GPS", "CONSOLA"} {
		add("tetra_viewer tabs", label, 179-12, 22, simpleui.FontRegular)
	}

	// tetra_panel.go drawTETRAStatusField puts the value 58 pixels along.
	for _, label := range []string{"ESTADO", "NIVEL", "CALIDAD", "AFC", "CENTRO"} {
		add("tetra_panel status fields", label, 56, 11, simpleui.FontSemiBold)
	}

	// main_screen.go bottom settings strip.
	for _, label := range []string{"ESTILO  OSCURO", "ESTILO  CLARO"} {
		add("main_screen theme button", label, 190-12, 13, simpleui.FontSemiBold)
	}

	// utilities_sidebar.go composes these from a translated prefix.
	for _, label := range []string{"REANUDAR AUTO", "REANUDAR DELAY", "REANUDAR HOLD"} {
		add("utilities_sidebar scan mode", label, 150-10, 11, simpleui.FontSemiBold)
	}
	for _, state := range []string{"PREPARADO", "BUSCANDO TRANSMISIONES", "ESCUCHA RETENIDA", "VERIFICANDO", "ESCUCHANDO", "CAMBIO"} {
		add("utilities_sidebar scan state", state, 350-24-12, 12, simpleui.FontSemiBold)
	}

	// recorder_panel.go format button.
	for _, label := range []string{"FORMATO MP3", "FORMATO WAV"} {
		add("recorder_panel format", label, 150-10, 12, simpleui.FontSemiBold)
	}

	// scan_panel.go drawScanButton: title at 14 and detail at 12, both centred
	// inside the button.
	for _, button := range []struct {
		width  float32
		title  string
		detail string
	}{
		{220, "AJUSTAR A MEMORIA", "OFF · Sintonizar el pico"},
		{220, "AJUSTAR A MEMORIA", "ON · Si coincide con un canal"},
		{220, "AL PERDER LA SEÑAL  ▾", "Continuar rápidamente"},
		{220, "AL PERDER LA SEÑAL  ▾", "Permanecer detenido"},
		{220, "AL PERDER LA SEÑAL  ▾", "Esperar y continuar"},
		{190, "ESPERA 3 s  ▾", "Antes de continuar"},
		{220, "INICIAR ESCANEO", "Buscar entre MIN y MAX"},
		{220, "DETENER ESCANEO", "Conservar frecuencia actual"},
		{220, "DURANTE LA ESCUCHA  ▾", "Saltar a otra más fuerte"},
		{220, "DURANTE LA ESCUCHA  ▾", "Mantener señal actual"},
	} {
		add("scan_panel button title", button.title, button.width-10, 14, simpleui.FontSemiBold)
		add("scan_panel button detail", button.detail, button.width-10, 12, simpleui.FontRegular)
	}

	return found
}
