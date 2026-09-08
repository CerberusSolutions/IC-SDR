package screens

import (
	"encoding/binary"
	"os"
	"testing"
	"time"
)

func TestAudioRecorderWritesValidWAV(t *testing.T) {
	recorder := newAudioRecorder(t.TempDir())
	recorder.Configure(446_093_750, "PMR446", "NFM")
	recorder.Start()
	recorder.Submit([]float32{0, .25, -.25, .5, -.5}, false, true)
	recorder.Stop()
	deadline := time.Now().Add(time.Second)
	for len(recorder.State().RecentFiles) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	state := recorder.State()
	if len(state.RecentFiles) != 1 {
		recorder.Close()
		t.Fatal("recording was not finalized")
	}
	data, err := os.ReadFile(state.RecentFiles[0])
	if err != nil {
		recorder.Close()
		t.Fatal(err)
	}
	if len(data) != 44+10 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		recorder.Close()
		t.Fatalf("invalid WAV: bytes=%d", len(data))
	}
	if got := binary.LittleEndian.Uint32(data[24:28]); got != audioSampleRate {
		recorder.Close()
		t.Fatalf("sample rate = %d", got)
	}
	if err := recorder.DeleteFile(state.RecentFiles[0]); err != nil {
		recorder.Close()
		t.Fatalf("delete recording: %v", err)
	}
	if _, err := os.Stat(state.RecentFiles[0]); !os.IsNotExist(err) {
		recorder.Close()
		t.Fatalf("recording still exists after delete: %v", err)
	}
	if len(recorder.State().RecentFiles) != 0 {
		recorder.Close()
		t.Fatal("deleted recording remains in recent list")
	}
	recorder.Close()
}

func TestAudioRecorderCanSkipClosedSquelch(t *testing.T) {
	recorder := newAudioRecorder(t.TempDir())
	recorder.SetSkipSquelchSilence(true)
	recorder.Start()
	recorder.Submit(make([]float32, 100), true, false)
	state := recorder.State()
	if state.DurationSeconds != 0 || !state.WaitingForSquelch {
		recorder.Close()
		t.Fatalf("closed squelch state = %#v", state)
	}
	recorder.Close()
}
