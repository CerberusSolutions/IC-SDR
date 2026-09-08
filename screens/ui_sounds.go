package screens

import (
	_ "embed"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/sounds/tool-select.wav
var toolSelectWAV []byte

//go:embed assets/sounds/button-click.wav
var buttonClickWAV []byte

type UISounds struct {
	loaded             bool
	toolSelect, button rl.Sound
}

func (sounds *UISounds) EnsureLoaded() {
	if sounds.loaded || !rl.IsAudioDeviceReady() {
		return
	}
	sounds.toolSelect = loadEmbeddedSound(toolSelectWAV)
	sounds.button = loadEmbeddedSound(buttonClickWAV)
	sounds.loaded = sounds.toolSelect.FrameCount > 0 && sounds.button.FrameCount > 0
}

func loadEmbeddedSound(data []byte) rl.Sound {
	if len(data) == 0 {
		return rl.Sound{}
	}
	wave := rl.LoadWaveFromMemory(".wav", data, int32(len(data)))
	if wave.FrameCount == 0 {
		return rl.Sound{}
	}
	sound := rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)
	return sound
}

func (sounds *UISounds) PlayToolSelect() {
	sounds.EnsureLoaded()
	if sounds.loaded {
		rl.PlaySound(sounds.toolSelect)
	}
}

func (sounds *UISounds) PlayButton() {
	sounds.EnsureLoaded()
	if sounds.loaded {
		rl.PlaySound(sounds.button)
	}
}

func (sounds *UISounds) Close() {
	if !sounds.loaded {
		return
	}
	rl.UnloadSound(sounds.toolSelect)
	rl.UnloadSound(sounds.button)
	sounds.loaded = false
}
