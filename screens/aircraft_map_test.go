package screens

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"go-zero/internal/aircraft"
	"go-zero/simpleui"
)

func TestAircraftMapProjectionRoundTrip(t *testing.T) {
	v := aircraftMap{centerLat: 40, centerLon: -4, lonSpan: 12}
	b := rl.Rectangle{X: 20, Y: 78, Width: 970, Height: 690}
	p := v.project(41.2, -2.3, b)
	lat, lon := v.unproject(p, b)
	if abs64(lat-41.2) > 1e-5 || abs64(lon+2.3) > 1e-5 {
		t.Fatalf("projection mismatch %.6f %.6f", lat, lon)
	}
}

func TestAircraftMapLabelClickAndZoomFromDetailsPanel(t *testing.T) {
	lat, lon := 40.48, -3.57
	v := aircraftMap{centerLat: lat, centerLon: lon, lonSpan: 2, selected: -1, list: []aircraft.Aircraft{{ICAO: "3451A2", Latitude: &lat, Longitude: &lon}}}
	b := rl.Rectangle{X: 20, Y: 78, Width: 970, Height: 690}
	p := v.project(lat, lon, b)
	v.selectAt(rl.Vector2{X: p.X + 120, Y: p.Y}, b)
	if v.selected != 0 {
		t.Fatal("aircraft label was not clickable")
	}
	before := v.lonSpan
	v.zoomAt(rl.Vector2{X: 1200, Y: 400}, b, 1)
	if v.lonSpan >= before {
		t.Fatal("wheel over details panel did not zoom aircraft map")
	}
}

func TestAircraftMapRender(t *testing.T) {
	dir := os.Getenv("AIRCRAFT_MAP_RENDER_DIR")
	if dir == "" {
		t.Skip("visual QA not requested")
	}
	_ = os.MkdirAll(dir, 0755)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(1360, 800, "Aircraft QA")
	defer rl.CloseWindow()
	simpleui.SetTextScale(1.25)
	lat, lon, spd, trk, vr := 40.48, -3.57, 265., 215., -640.
	alt := 11800
	list := []aircraft.Aircraft{{ICAO: "3451A2", Callsign: "IBE3174", Latitude: &lat, Longitude: &lon, Altitude: &alt, Speed: &spd, Track: &trk, VerticalRate: &vr, Source: aircraft.Mode1090, LastSeen: time.Now(), Messages: 284}}
	data, _ := json.Marshal(list)
	path := filepath.Join(dir, "aircraft.json")
	_ = os.WriteFile(path, data, 0644)
	v := &aircraftMap{path: path, centerLat: 40.4, centerLon: -3.7, lonSpan: 4, selected: 0, tracks: map[string][]geoPoint{"3451A2": {{40.9, -2.8}, {40.7, -3.1}, {40.48, -3.57}}}}
	canvas := rl.LoadRenderTexture(1360, 800)
	defer rl.UnloadRenderTexture(canvas)
	rl.BeginTextureMode(canvas)
	v.draw()
	rl.EndTextureMode()
	img := rl.LoadImageFromTexture(canvas.Texture)
	defer rl.UnloadImage(img)
	rl.ImageFlipVertical(img)
	if !rl.ExportImage(*img, filepath.Join(dir, "aircraft-map.png")) {
		t.Fatal("render failed")
	}
}
