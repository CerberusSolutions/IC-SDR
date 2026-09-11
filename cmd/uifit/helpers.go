package main

import (
	"go/ast"
	"go/token"
	"strconv"
)

// Most panels wrap simpleui.NewButton in a small local helper so they can apply
// a house style in one place:
//
//	button := func(id, label string, x, w float32, action func()) *simpleui.Button {
//	    b := simpleui.NewButton(id, x, toolY+28, w, 34, label, uiControlFontSize)
//
// A scanner that only recognises simpleui.NewButton measures none of those
// labels, and would happily report that every string fits while never looking
// at the buttons that make up most of each decoder panel.
//
// Rather than hard-code each helper's signature, resolve it: find the
// NewButton call inside the helper and see which of the helper's own parameters
// supply the label, the width and the font size. A helper that passes a
// constant for one of them is resolved through constants too.
type buttonHelper struct {
	labelParam int // index into the helper's parameter list
	widthParam int
	fontSize   int32
}

// resolveHelpers walks the files of one package and returns the helpers it can
// resolve, keyed by the name they are called through.
func resolveHelpers(files map[string]*ast.File, constants map[string]int32) map[string]buttonHelper {
	helpers := map[string]buttonHelper{}
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			name, params, body := helperParts(node)
			if body == nil || name == "" {
				return true
			}
			if helper, ok := resolveHelper(params, body, constants); ok {
				helpers[name] = helper
			}
			return true
		})
	}
	return helpers
}

// helperParts recognises both shapes a helper takes: a method on the panel, and
// a function literal assigned to a local name.
func helperParts(node ast.Node) (string, []string, *ast.BlockStmt) {
	switch declaration := node.(type) {
	case *ast.FuncDecl:
		if declaration.Body == nil {
			return "", nil, nil
		}
		return declaration.Name.Name, parameterNames(declaration.Type), declaration.Body
	case *ast.AssignStmt:
		if len(declaration.Lhs) != 1 || len(declaration.Rhs) != 1 {
			return "", nil, nil
		}
		name, ok := declaration.Lhs[0].(*ast.Ident)
		if !ok {
			return "", nil, nil
		}
		literal, ok := declaration.Rhs[0].(*ast.FuncLit)
		if !ok {
			return "", nil, nil
		}
		return name.Name, parameterNames(literal.Type), literal.Body
	}
	return "", nil, nil
}

func parameterNames(signature *ast.FuncType) []string {
	var names []string
	if signature.Params == nil {
		return names
	}
	for _, field := range signature.Params.List {
		if len(field.Names) == 0 {
			names = append(names, "")
			continue
		}
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

// resolveHelper finds a simpleui.NewButton call in body and maps its label,
// width and font size back to the helper's parameters.
func resolveHelper(params []string, body *ast.BlockStmt, constants map[string]int32) (buttonHelper, bool) {
	index := func(name string) int {
		for i, param := range params {
			if param == name {
				return i
			}
		}
		return -1
	}

	helper := buttonHelper{labelParam: -1, widthParam: -1}
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 7 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "NewButton" {
			return true
		}
		if pkg, ok := selector.X.(*ast.Ident); !ok || pkg.Name != "simpleui" {
			return true
		}
		// NewButton(id, x, y, width, height, label, fontSize)
		if width, ok := call.Args[3].(*ast.Ident); ok {
			helper.widthParam = index(width.Name)
		}
		if label, ok := call.Args[5].(*ast.Ident); ok {
			helper.labelParam = index(label.Name)
		}
		switch size := call.Args[6].(type) {
		case *ast.Ident:
			if value, known := constants[size.Name]; known {
				helper.fontSize = value
			}
		case *ast.BasicLit:
			if size.Kind == token.INT {
				if value, err := strconv.Atoi(size.Value); err == nil {
					helper.fontSize = int32(value)
				}
			}
		}
		found = helper.widthParam >= 0 && helper.labelParam >= 0 && helper.fontSize > 0
		return !found
	})
	return helper, found
}

// packageConstants collects the int32 constants a helper might use for its font
// size, such as uiControlFontSize.
func packageConstants(files map[string]*ast.File) map[string]int32 {
	constants := map[string]int32{}
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.ValueSpec)
			if !ok || len(spec.Names) != len(spec.Values) {
				return true
			}
			for i, name := range spec.Names {
				call, ok := spec.Values[i].(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					continue
				}
				if kind, ok := call.Fun.(*ast.Ident); !ok || kind.Name != "int32" {
					continue
				}
				literal, ok := call.Args[0].(*ast.BasicLit)
				if !ok || literal.Kind != token.INT {
					continue
				}
				if value, err := strconv.Atoi(literal.Value); err == nil {
					constants[name.Name] = int32(value)
				}
			}
			return true
		})
	}
	return constants
}
