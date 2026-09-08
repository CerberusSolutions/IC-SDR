package tetra

import "testing"

func appendBits(dst []byte, value uint32, width int) []byte {
	for i := width - 1; i >= 0; i-- {
		dst = append(dst, byte(value>>uint(i))&1)
	}
	return dst
}

func TestParseNeighbourBroadcast(t *testing.T) {
	var bits []byte
	bits = appendBits(bits, 5, 3) // MLE
	bits = appendBits(bits, 2, 3) // D-NWRK-BROADCAST
	bits = appendBits(bits, 0, 16)
	bits = appendBits(bits, 2, 2)
	bits = appendBits(bits, 1, 1) // options
	bits = appendBits(bits, 0, 1) // no network time
	bits = appendBits(bits, 1, 1) // neighbour count present
	bits = appendBits(bits, 1, 3)
	bits = appendBits(bits, 7, 5)
	bits = appendBits(bits, 0, 2)
	bits = appendBits(bits, 1, 1)
	bits = appendBits(bits, 3, 2)
	bits = appendBits(bits, 3666, 12)
	bits = appendBits(bits, 0, 1) // no neighbour options
	network := NetworkInfo{Valid: true, FrequencyBand: 3, FrequencyOffset: 0}
	cells, ok := parseNeighbourBroadcast(bits, SystemInfo{MCC: 214, MNC: 8}, network)
	if !ok || len(cells) != 1 {
		t.Fatalf("cells=%#v ok=%v", cells, ok)
	}
	cell := cells[0]
	if cell.CellID != 7 || cell.Carrier != 3666 || cell.FrequencyHz != 391_650_000 || !cell.Synchronized || cell.ServiceLevel != 3 {
		t.Fatalf("unexpected neighbour: %#v", cell)
	}
}
