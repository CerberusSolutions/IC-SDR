package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"go-zero/simpleui"
)

// scanSource reads the screens package and pulls out every string whose
// available width is written in the source as a literal: control constructors,
// centred text in a literal rectangle, and the column header tables the data
// panels use.
//
// It parses rather than imports, so the tool builds and runs on any platform
// even though the screens package itself is Windows-only.
func scanSource(root string) ([]measurement, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, filepath.Join(root, "screens"), func(entry fs.FileInfo) bool {
		return !strings.HasSuffix(entry.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	var found []measurement
	for _, pkg := range packages {
		// Panels wrap NewButton in local helpers; resolve those first so their
		// labels are measured alongside the direct calls.
		constants := packageConstants(pkg.Files)
		helpers := resolveHelpers(pkg.Files, constants)
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				where := func(n ast.Node) string {
					position := fset.Position(n.Pos())
					return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
				}
				if call, ok := node.(*ast.CallExpr); ok {
					found = append(found, fromConstructor(call, where(call), constants)...)
					found = append(found, fromHelper(call, helpers, where(call))...)
				}
				if composite, ok := node.(*ast.CompositeLit); ok {
					found = append(found, fromColumnTable(composite, where(composite))...)
				}
				return true
			})
		}
	}
	return found, nil
}

// fromConstructor recognises the SimpleUI control constructors and the screens
// package's own centred-text helpers.
// The workspace panels were laid out for the former full-width window and are
// squeezed into the area beside the utilities column at run time.
// compactToolControls scales their width, but SimpleUI deliberately leaves text
// measurement alone, so a label in that area has proportionally less room than
// its declared width suggests.
const (
	toolY             = 630
	toolH             = 196
	toolContentScaleX = (1592 - 360) / 1552.0
)

// available returns the room a control really has, given where it sits.
func available(declared float64, y float64, padding float32) float32 {
	width := float32(declared)
	if y >= toolY && y < toolY+toolH {
		width *= toolContentScaleX
	}
	return width - padding
}

func fromConstructor(call *ast.CallExpr, where string, constants map[string]int32) []measurement {
	name := ""
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "simpleui" {
			name = fun.Sel.Name
		}
	case *ast.Ident:
		name = fun.Name
	}

	switch name {
	case "NewButton": // id, x, y, width, height, label, fontSize
		if len(call.Args) != 7 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		y, okY := number(call.Args[2])
		label, ok2 := text(call.Args[5])
		size, ok3 := fontSize(call.Args[6], constants)
		if ok1 && ok2 && ok3 && okY {
			// Buttons centre their label with no padding of their own; leave a
			// few pixels either side so the text never touches the border.
			return one(where, "button", label, available(width, y, 12), size, simpleui.FontSemiBold)
		}
	case "NewLabel": // id, x, y, width, height, text, fontSize
		if len(call.Args) != 7 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		y, okY := number(call.Args[2])
		label, ok2 := text(call.Args[5])
		size, ok3 := fontSize(call.Args[6], constants)
		if ok1 && ok2 && ok3 && okY && label != "" {
			return one(where, "label", label, available(width, y, 0), size, simpleui.FontRegular)
		}
	case "NewSwitch": // id, x, y, width, height, label, active, fontSize
		if len(call.Args) != 8 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		height, ok2 := number(call.Args[4])
		y, okY := number(call.Args[2])
		label, ok3 := text(call.Args[5])
		size, ok4 := fontSize(call.Args[7], constants)
		if ok1 && ok2 && ok3 && ok4 && okY {
			// The label is left-aligned and the track sits hard against the
			// right edge; the two must not meet. This mirrors switch.trackBounds.
			drawn := available(width, y, 0)
			track := min(float32(height)*0.68, 26)
			track = min(max(track*1.8, 42), drawn*0.45)
			return one(where, "switch", label, drawn-track-6, size, simpleui.FontSemiBold)
		}
	case "NewDropdown": // id, x, y, width, height, selected, items, fontSize
		if len(call.Args) != 8 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		y, okY := number(call.Args[2])
		label, ok2 := text(call.Args[5])
		size, ok3 := fontSize(call.Args[7], constants)
		if ok1 && ok2 && ok3 && okY {
			// 12px indent plus room for the chevron.
			return one(where, "dropdown", label, available(width, y, 34), size, simpleui.FontSemiBold)
		}
	case "drawCentered", "drawCenteredStyled": // text, rectangle, size, [style], colour
		if len(call.Args) < 4 {
			return nil
		}
		label, ok := text(call.Args[0])
		if !ok {
			return nil
		}
		width, ok := rectangleWidth(call.Args[1])
		if !ok {
			return nil
		}
		size, ok := number(call.Args[2])
		if !ok {
			return nil
		}
		style := simpleui.FontRegular
		if name == "drawCenteredStyled" {
			style = simpleui.FontSemiBold
		}
		return one(where, "centred", label, width-8, int32(size), style)
	}
	return nil
}

// fromColumnTable recognises the {x, "HEADER"} tables the data panels use for
// their column headings. Each heading has until the next column starts.
func fromColumnTable(composite *ast.CompositeLit, where string) []measurement {
	type column struct {
		x     float64
		label string
	}
	var columns []column
	for _, element := range composite.Elts {
		inner, ok := element.(*ast.CompositeLit)
		if !ok || len(inner.Elts) != 2 {
			return nil
		}
		x, ok1 := number(inner.Elts[0])
		label, ok2 := text(inner.Elts[1])
		if !ok1 || !ok2 {
			return nil
		}
		columns = append(columns, column{x, label})
	}
	if len(columns) < 3 {
		return nil
	}
	for i := 1; i < len(columns); i++ {
		if columns[i].x <= columns[i-1].x {
			return nil
		}
	}
	var found []measurement
	for i, c := range columns {
		// The last column runs to the edge of its panel; allow it the width of
		// a typical column rather than guessing at the panel.
		available := float32(120)
		if i+1 < len(columns) {
			available = float32(columns[i+1].x-c.x) - 8
		}
		found = append(found, measurement{where, "column", c.label, available, 12, simpleui.FontSemiBold})
	}
	return found
}

func one(where, kind, label string, available float32, size int32, style simpleui.FontStyle) []measurement {
	return []measurement{{where, kind, label, available, size, style}}
}

// fontSize resolves a size written either as a literal or as one of the
// package's named constants, such as uiControlFontSize.
func fontSize(expr ast.Expr, constants map[string]int32) (int32, bool) {
	if value, ok := number(expr); ok {
		return int32(value), true
	}
	if name, ok := expr.(*ast.Ident); ok {
		if value, known := constants[name.Name]; known {
			return value, true
		}
	}
	return 0, false
}

func text(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

// number evaluates the constant arithmetic the layout is written in. Positions
// are routinely expressed against the workspace constants, as in toolY+28, so
// a plain literal check would skip most of the controls in the decoder panels.
func number(expr ast.Expr) (float64, bool) { return evaluate(expr, nil) }

func evaluate(expr ast.Expr, constants map[string]int32) (float64, bool) {
	switch node := expr.(type) {
	case *ast.BasicLit:
		if node.Kind == token.INT || node.Kind == token.FLOAT {
			value, err := strconv.ParseFloat(node.Value, 64)
			return value, err == nil
		}
	case *ast.Ident:
		if value, known := layoutConstants[node.Name]; known {
			return value, true
		}
		if value, known := constants[node.Name]; known {
			return float64(value), true
		}
	case *ast.ParenExpr:
		return evaluate(node.X, constants)
	case *ast.UnaryExpr:
		if node.Op == token.SUB {
			if value, ok := evaluate(node.X, constants); ok {
				return -value, true
			}
		}
	case *ast.BinaryExpr:
		left, okL := evaluate(node.X, constants)
		right, okR := evaluate(node.Y, constants)
		if !okL || !okR {
			return 0, false
		}
		switch node.Op {
		case token.ADD:
			return left + right, true
		case token.SUB:
			return left - right, true
		case token.MUL:
			return left * right, true
		case token.QUO:
			if right != 0 {
				return left / right, true
			}
		}
	case *ast.CallExpr: // float32(240)
		if len(node.Args) == 1 {
			return evaluate(node.Args[0], constants)
		}
	}
	return 0, false
}

// layoutConstants are the workspace coordinates the panels are written against.
// They mirror the values in screens/main_screen.go.
var layoutConstants = map[string]float64{
	"designWidth": 1600, "designHeight": 900,
	"waterfallY": 450, "waterfallH": 170,
	"toolY": toolY, "toolH": toolH,
	"legacyToolX": 24, "legacyToolWidth": 1552,
	"toolContentX": 360, "toolContentRight": 1592,
	"utilitiesRight": 350,
}

func rectangleWidth(expr ast.Expr) (float32, bool) {
	composite, ok := expr.(*ast.CompositeLit)
	if !ok {
		return 0, false
	}
	for _, element := range composite.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := pair.Key.(*ast.Ident); ok && key.Name == "Width" {
			if width, ok := number(pair.Value); ok {
				return float32(width), true
			}
		}
	}
	return 0, false
}

// fromHelper measures a label passed through one of the panels' own button
// helpers, using the label and width positions resolved from the helper body.
func fromHelper(call *ast.CallExpr, helpers map[string]buttonHelper, where string) []measurement {
	name := ""
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		name = fun.Name
	case *ast.SelectorExpr:
		// A method helper such as p.button.
		name = fun.Sel.Name
	}
	helper, known := helpers[name]
	if !known {
		return nil
	}
	if helper.labelParam >= len(call.Args) || helper.widthParam >= len(call.Args) {
		return nil
	}
	label, ok := text(call.Args[helper.labelParam])
	if !ok {
		return nil
	}
	width, ok := number(call.Args[helper.widthParam])
	if !ok {
		return nil
	}
	// Helper-built buttons sit in the workspace, so they are squeezed too.
	return one(where, "button", label, available(width, toolY, 12), helper.fontSize, simpleui.FontSemiBold)
}
