package screens

import rl "github.com/gen2brain/raylib-go/raylib"

// Flags are drawn from primitives rather than shipped as bitmaps. At the size
// the language selector uses them a 24-pixel-tall PNG would either alias badly
// or need several resolutions, and both flags reduce to a handful of
// rectangles and triangles.

var (
	// Spain: the shades used by the Bandera de España.
	spainRed    = rl.Color{R: 170, G: 21, B: 27, A: 255}
	spainYellow = rl.Color{R: 241, G: 191, B: 0, A: 255}

	// United Kingdom: Pantone 280 C and 186 C as rendered in sRGB.
	unionBlue  = rl.Color{R: 1, G: 33, B: 105, A: 255}
	unionRed   = rl.Color{R: 200, G: 16, B: 46, A: 255}
	unionWhite = rl.Color{R: 255, G: 255, B: 255, A: 255}
)

// drawSpanishFlag paints the Spanish flag inside the given rectangle. The
// horizontal bands are a quarter, a half and a quarter of the height; the coat
// of arms is omitted because it is illegible at this size.
func drawSpanishFlag(bounds rl.Rectangle) {
	band := bounds.Height / 4
	rl.DrawRectangleRec(rl.Rectangle{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: band}, spainRed)
	rl.DrawRectangleRec(rl.Rectangle{X: bounds.X, Y: bounds.Y + band, Width: bounds.Width, Height: band * 2}, spainYellow)
	rl.DrawRectangleRec(rl.Rectangle{X: bounds.X, Y: bounds.Y + band*3, Width: bounds.Width, Height: bounds.Height - band*3}, spainRed)
}

// drawUnionFlag paints the Union Flag inside the given rectangle: the blue
// field, the white saltire of St Andrew with the red saltire of St Patrick
// counterchanged within it, then the white-fimbriated red cross of St George.
func drawUnionFlag(bounds rl.Rectangle) {
	x, y, w, h := bounds.X, bounds.Y, bounds.Width, bounds.Height
	rl.DrawRectangleRec(bounds, unionBlue)

	// Arm half-widths, given per axis so the saltire meets each corner square
	// on whatever aspect ratio the flag is drawn at.
	dx, dy := w*0.15, h*0.15

	// St Andrew's white saltire: one band per diagonal, centred on the
	// diagonal and clipped by the edges of the flag, which is why each end is
	// a six-sided figure rather than a simple parallelogram. Drawing it as a
	// parallelogram is the usual mistake and leaves blue notches in the
	// corners.
	fan(unionWhite,
		x, y, x+dx, y, x+w, y+h-dy, x+w, y+h, x+w-dx, y+h, x, y+dy)
	fan(unionWhite,
		x, y+h, x, y+h-dy, x+w-dx, y, x+w, y, x+w, y+dy, x+dx, y+h)

	// St Patrick's red saltire sits inside the white and is counterchanged:
	// in the hoist half of each diagonal the red hugs the lower edge of the
	// white arm, and in the fly half it hugs the upper edge. That switch at
	// the centre is what gives the Union Flag its familiar asymmetry.
	const red = 0.45
	drawCounterchangedArm(x, y+dy, x+dx, y, x+w, y+h-dy, x+w-dx, y+h, red)
	drawCounterchangedArm(x, y+h-dy, x+dx, y+h, x+w, y+dy, x+w-dx, y, red)

	// St George's cross, fimbriated in white.
	crossW, crossH := w*0.30, h*0.30
	rl.DrawRectangleRec(rl.Rectangle{X: x + (w-crossW)/2, Y: y, Width: crossW, Height: h}, unionWhite)
	rl.DrawRectangleRec(rl.Rectangle{X: x, Y: y + (h-crossH)/2, Width: w, Height: crossH}, unionWhite)
	innerW, innerH := w*0.18, h*0.18
	rl.DrawRectangleRec(rl.Rectangle{X: x + (w-innerW)/2, Y: y, Width: innerW, Height: h}, unionRed)
	rl.DrawRectangleRec(rl.Rectangle{X: x, Y: y + (h-innerH)/2, Width: w, Height: innerH}, unionRed)
}

// drawCounterchangedArm fills the red part of one white saltire band. The band
// is given as the quadrilateral a-b-c-d, where a→d is one long edge and b→c is
// the other; ratio is how much of the band's width the red takes.
//
// The band is split at its midpoint: the first half takes its colour from the
// a-d edge and the second half from the b-c edge, which is the counterchange.
func drawCounterchangedArm(ax, ay, bx, by, cx, cy, dx, dy, ratio float32) {
	// Perpendicular step from the a-d edge across to the b-c edge.
	stepX, stepY := (bx-ax)*ratio, (by-ay)*ratio
	// Midpoints of each long edge.
	mADx, mADy := (ax+dx)/2, (ay+dy)/2
	mBCx, mBCy := (bx+cx)/2, (by+cy)/2

	// Hoist half: red along the a-d edge.
	quad(ax, ay, ax+stepX, ay+stepY, mADx+stepX, mADy+stepY, mADx, mADy, unionRed)
	// Fly half: red along the b-c edge.
	quad(mBCx, mBCy, mBCx-stepX, mBCy-stepY, cx-stepX, cy-stepY, cx, cy, unionRed)
}

// fan fills a convex polygon given as consecutive x, y pairs in order around
// its perimeter.
func fan(color rl.Color, points ...float32) {
	vertices := make([]rl.Vector2, 0, len(points)/2)
	for i := 0; i+1 < len(points); i += 2 {
		vertices = append(vertices, rl.Vector2{X: points[i], Y: points[i+1]})
	}
	for i := 1; i+1 < len(vertices); i++ {
		rl.DrawTriangle(vertices[0], vertices[i], vertices[i+1], color)
		rl.DrawTriangle(vertices[0], vertices[i+1], vertices[i], color)
	}
}

// quad fills the convex quadrilateral a-b-c-d, given in order around its
// perimeter, as two triangles.
func quad(ax, ay, bx, by, cx, cy, dx, dy float32, color rl.Color) {
	a := rl.Vector2{X: ax, Y: ay}
	b := rl.Vector2{X: bx, Y: by}
	c := rl.Vector2{X: cx, Y: cy}
	d := rl.Vector2{X: dx, Y: dy}
	// Raylib expects counter-clockwise winding; emitting both orders keeps the
	// shape visible whichever way round the caller listed its corners.
	rl.DrawTriangle(a, b, c, color)
	rl.DrawTriangle(a, c, b, color)
	rl.DrawTriangle(a, c, d, color)
	rl.DrawTriangle(a, d, c, color)
}
