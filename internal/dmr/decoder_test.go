package dmr

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"
)

func TestCandidateAudioIsReleasedAfterConfirmation(t *testing.T) {
	var got []float32
	d := New(2_048_000, "", func(samples []float32) { got = append(got, samples...) })
	d.enabled.Store(true)
	d.generation = 1
	encode := func(values ...int16) []byte {
		var buffer bytes.Buffer
		if err := binary.Write(&buffer, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		return buffer.Bytes()
	}
	d.audioLoop(bytes.NewReader(encode(1000, 2000)), 1)
	if len(got) != 0 || len(d.pendingAudio) != 2 {
		t.Fatalf("candidate got=%d pending=%d", len(got), len(d.pendingAudio))
	}
	d.confirmed = true
	d.audioLoop(bytes.NewReader(encode(3000)), 1)
	if len(got) != 3 || len(d.pendingAudio) != 0 {
		t.Fatalf("confirmed got=%d pending=%d", len(got), len(d.pendingAudio))
	}
}

type fragmentedReader struct {
	data  []byte
	sizes []int
	read  int
}

func TestDMRStateRequiresStableTelemetry(t *testing.T) {
	d := New(2_048_000, "", nil)
	start := time.Now()
	d.mu.Lock()
	d.handleRawStateLocked("VOICE", start)
	if d.status.State != "CANDIDATE" || d.confirmed {
		t.Fatalf("first sync state=%s confirmed=%v", d.status.State, d.confirmed)
	}
	d.status.PLLLocked, d.status.SyncQuality, d.status.InputLevel = true, 5, 10
	d.updateConfirmationLocked(start.Add(100 * time.Millisecond))
	if d.confirmed {
		t.Fatal("DMR confirmed before temporal guard")
	}
	d.updateConfirmationLocked(start.Add(250 * time.Millisecond))
	if !d.confirmed || d.status.State != "VOICE" {
		t.Fatalf("stable DMR not confirmed: state=%s confirmed=%v", d.status.State, d.confirmed)
	}
	d.mu.Unlock()
}

func (r *fragmentedReader) Read(target []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	size := r.sizes[r.read%len(r.sizes)]
	r.read++
	size = min(size, min(len(r.data), len(target)))
	copy(target, r.data[:size])
	r.data = r.data[size:]
	return size, nil
}

func TestDecodePCM16LEPreservesOddPipeFragments(t *testing.T) {
	raw := []byte{0x00, 0x80, 0xff, 0xff, 0x00, 0x00, 0xff, 0x7f}
	reader := &fragmentedReader{data: bytes.Clone(raw), sizes: []int{1, 3, 1, 2, 1}}
	var got []float32
	if err := decodePCM16LE(reader, func(samples []float32) { got = append(got, samples...) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("samples=%d, want 4", len(got))
	}
	want := []float32{-1, -1.0 / 32768, 0, 32767.0 / 32768}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sample[%d]=%v, want %v", i, got[i], want[i])
		}
	}
}
