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

func TestEditGroupRenamesAndAppliesSharedFlags(t *testing.T) {
	dir := t.TempDir()
	panel := &MemoryPanel{
		memories:    []MemoryEntry{{Name: "A", Group: "OLD"}, {Name: "B", Group: "OLD", Priority: true}},
		groupColors: map[string]string{"OLD": "#000000"}, groupColorsPath: filepath.Join(dir, "groups.json"), path: filepath.Join(dir, "memories.json"),
		selectedGroup: "OLD", pendingOriginalGroup: "OLD", pendingGroup: "NEW", pendingGroupColor: 2, pendingGroupScan: true,
	}
	panel.commitGroupColor()
	if panel.selectedGroup != "NEW" || panel.memories[0].Group != "NEW" || panel.memories[1].Group != "NEW" || !panel.memories[0].ScanEnabled || panel.memories[1].Priority {
		t.Fatalf("group edit was not applied: %+v", panel)
	}
	if _, old := panel.groupColors["OLD"]; old || panel.groupColor("NEW") != memoryGroupPalette[2] {
		t.Fatalf("group color was not moved: %+v", panel.groupColors)
	}
}

func TestEditGroupCanReturnToNoMove(t *testing.T) {
	panel := &MemoryPanel{
		groups:               []string{"TODAS", "OLD", "DESTINO"},
		pendingOriginalGroup: "OLD",
	}
	panel.moveEditedGroup()
	if panel.pendingMoveGroup == "" {
		t.Fatal("el primer clic debe seleccionar un destino")
	}
	for panel.pendingMoveGroup != "" {
		panel.moveEditedGroup()
	}
	if panel.pendingMoveGroup != "" {
		t.Fatal("el selector debe poder volver a NO MOVER")
	}
}

func TestEditGroupMovesMemoriesIntoExistingGroup(t *testing.T) {
	dir := t.TempDir()
	panel := &MemoryPanel{
		memories: []MemoryEntry{
			{Name: "A", Group: "OLD"},
			{Name: "B", Group: "DESTINO"},
		},
		groupColors:          map[string]string{"OLD": "#000000", "DESTINO": "#123456"},
		groupColorsPath:      filepath.Join(dir, "groups.json"),
		path:                 filepath.Join(dir, "memories.json"),
		selectedGroup:        "OLD",
		pendingOriginalGroup: "OLD",
		pendingGroup:         "OLD",
		pendingMoveGroup:     "DESTINO",
		pendingGroupScan:     true,
	}
	panel.commitGroupColor()
	if panel.memories[0].Group != "DESTINO" || panel.selectedGroup != "DESTINO" {
		t.Fatalf("las memorias no se movieron al grupo existente: %+v", panel)
	}
	if panel.groupColors["DESTINO"] != "#123456" {
		t.Fatal("mover memorias no debe sustituir el color del grupo de destino")
	}
}
