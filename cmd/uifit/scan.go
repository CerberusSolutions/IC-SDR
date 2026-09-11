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
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				where := func(n ast.Node) string {
					position := fset.Position(n.Pos())
					return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
				}
				if call, ok := node.(*ast.CallExpr); ok {
					found = append(found, fromConstructor(call, where(call))...)
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
func fromConstructor(call *ast.CallExpr, where string) []measurement {
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
		label, ok2 := text(call.Args[5])
		size, ok3 := number(call.Args[6])
		if ok1 && ok2 && ok3 {
			// Buttons centre their label with no padding of their own; leave a
			// few pixels either side so the text never touches the border.
			return one(where, "button", label, float32(width)-12, int32(size), simpleui.FontSemiBold)
		}
	case "NewLabel": // id, x, y, width, height, text, fontSize
		if len(call.Args) != 7 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		label, ok2 := text(call.Args[5])
		size, ok3 := number(call.Args[6])
		if ok1 && ok2 && ok3 && label != "" {
			return one(where, "label", label, float32(width), int32(size), simpleui.FontRegular)
		}
	case "NewSwitch": // id, x, y, width, height, label, active, fontSize
		if len(call.Args) != 8 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		height, ok2 := number(call.Args[4])
		label, ok3 := text(call.Args[5])
		size, ok4 := number(call.Args[7])
		if ok1 && ok2 && ok3 && ok4 {
			// The label is left-aligned and the track sits hard against the
			// right edge; the two must not meet. This mirrors switch.trackBounds.
			track := min(float32(height)*0.68, 26)
			track = min(max(track*1.8, 42), float32(width)*0.45)
			return one(where, "switch", label, float32(width)-track-6, int32(size), simpleui.FontSemiBold)
		}
	case "NewDropdown": // id, x, y, width, height, selected, items, fontSize
		if len(call.Args) != 8 {
			return nil
		}
		width, ok1 := number(call.Args[3])
		label, ok2 := text(call.Args[5])
		size, ok3 := number(call.Args[7])
		if ok1 && ok2 && ok3 {
			// 12px indent plus room for the chevron.
			return one(where, "dropdown", label, float32(width)-34, int32(size), simpleui.FontSemiBold)
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

func text(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func number(expr ast.Expr) (float64, bool) {
	switch node := expr.(type) {
	case *ast.BasicLit:
		if node.Kind == token.INT || node.Kind == token.FLOAT {
			value, err := strconv.ParseFloat(node.Value, 64)
			return value, err == nil
		}
	case *ast.UnaryExpr:
		if node.Op == token.SUB {
			if value, ok := number(node.X); ok {
				return -value, true
			}
		}
	case *ast.CallExpr: // float32(240)
		if len(node.Args) == 1 {
			return number(node.Args[0])
		}
	}
	return 0, false
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
