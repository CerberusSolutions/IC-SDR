package screens

import (
	"path/filepath"
	"testing"
)

func TestMemoryGroupColorPersistsSeparately(t *testing.T) {
	path := filepath.Join(t.TempDir(), "groups.json")
	panel := &MemoryPanel{groupColorsPath: path, groupColors: map[string]string{}, pendingGroup: "DMR", pendingGroupColor: 3}
	panel.commitGroupColor()
	loaded := &MemoryPanel{groupColorsPath: path}
	loaded.loadGroupColors()
	color := loaded.groupColor("DMR")
	want := memoryGroupPalette[3]
	if color != want {
		t.Fatalf("color=%+v want %+v", color, want)
	}
}
func TestUnknownGroupKeepsDefaultCyan(t *testing.T) {
	panel := &MemoryPanel{groupColors: map[string]string{}}
	if got := panel.groupColor("NUEVO"); got != colors.cyan {
		t.Fatalf("fallback=%+v", got)
	}
}

func TestCreateEmptyMemoryGroupPersists(t *testing.T) {
	dir := t.TempDir()
	panel := &MemoryPanel{groupColorsPath: filepath.Join(dir, "groups.json"), groupColors: map[string]string{}, selectedGroup: "TODAS"}
	panel.rebuildGroups()
	panel.pendingGroup = "Emergencias"
	panel.pendingGroupColor = 3
	panel.commitNewGroup()
	if panel.selectedGroup != "Emergencias" || !panel.groupExists("emergencias") {
		t.Fatalf("new group was not selected or indexed: %+v", panel.groups)
	}
	loaded := &MemoryPanel{groupColorsPath: panel.groupColorsPath, groupColors: map[string]string{}, selectedGroup: "TODAS"}
	loaded.loadGroupColors()
	loaded.rebuildGroups()
	if !loaded.groupExists("EMERGENCIAS") {
		t.Fatalf("empty group did not survive reload: %+v", loaded.groups)
	}
}
