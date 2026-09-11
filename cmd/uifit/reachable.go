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

	"go-zero/internal/i18n"
)

// Interface text is normally translated on its way to the screen, which means
// the string SimpleUI is handed has to be a whole catalogue entry. A Spanish
// string that is joined to something else before it is drawn no longer matches
// anything, so it reaches the user untranslated.
//
// That is not visible in the catalogue, which still lists the entry, and it is
// not visible in the source, where the literal looks translated like any other.
// It only shows up in the running application — which is how the memory card's
// "PMR · SCAN ACTIVO" line was found, by eye, after the rest was done.
//
// This check finds them instead. It reports three shapes:
//
//	feedback += " · SOLO ABIERTA"        a catalogue entry appended to something
//	DrawText("BANDA " + name)            a catalogue entry joined inline
//	status := "SCAN ACTIVO"; ... a + status    joined by way of a variable
//
// Wrapping the piece in i18n.T at the point it is joined fixes all three.

type unreachable struct {
	where  string
	text   string
	reason string
}

func checkReachable(root string) ([]unreachable, error) {
	fset := token.NewFileSet()
	var found []unreachable

	dirs, err := interfaceDirs(root)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		packages, err := parser.ParseDir(fset, dir, func(entry fs.FileInfo) bool {
			return !strings.HasSuffix(entry.Name(), "_test.go")
		}, 0)
		if err != nil {
			return nil, err
		}
		for _, pkg := range packages {
			for _, file := range pkg.Files {
				found = append(found, checkFile(fset, file)...)
			}
		}
	}
	return found, nil
}

// interfaceDirs lists the package directories that carry interface text.
func interfaceDirs(root string) ([]string, error) {
	var dirs []string
	for _, base := range []string{"screens", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, base), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// The catalogue itself is full of catalogue entries by definition,
			// and uifit's own tables quote them deliberately.
			if entry.IsDir() && entry.Name() != "i18n" && entry.Name() != "uifit" {
				dirs = append(dirs, path)
			}
			if entry.IsDir() && (entry.Name() == "i18n" || entry.Name() == "uifit") {
				return fs.SkipDir
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return dirs, nil
}

func checkFile(fset *token.FileSet, file *ast.File) []unreachable {
	var found []unreachable
	at := func(node ast.Node) string {
		position := fset.Position(node.Pos())
		return fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line)
	}

	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			return true
		}

		// Names bound to a catalogue entry somewhere in this function.
		bound := map[string]string{}
		ast.Inspect(function.Body, func(inner ast.Node) bool {
			assign, ok := inner.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for index, value := range assign.Rhs {
				entry, ok := catalogueEntry(value)
				if !ok || index >= len(assign.Lhs) {
					continue
				}
				if name, ok := assign.Lhs[index].(*ast.Ident); ok {
					bound[name.Name] = entry
				}
			}
			return true
		})

		ast.Inspect(function.Body, func(inner ast.Node) bool {
			switch stmt := inner.(type) {
			case *ast.AssignStmt:
				// feedback += "..." — the appended part is never looked up.
				if stmt.Tok == token.ADD_ASSIGN {
					for _, value := range stmt.Rhs {
						if entry, ok := catalogueEntry(value); ok {
							found = append(found, unreachable{at(stmt), entry, "appended with +="})
						}
					}
				}
			case *ast.BinaryExpr:
				if stmt.Op != token.ADD {
					return true
				}
				for _, side := range []ast.Expr{stmt.X, stmt.Y} {
					if entry, ok := catalogueEntry(side); ok {
						found = append(found, unreachable{at(stmt), entry, "joined inline"})
						continue
					}
					if name, ok := side.(*ast.Ident); ok {
						if entry, bound := bound[name.Name]; bound {
							found = append(found, unreachable{at(stmt), entry, "joined by way of " + name.Name})
						}
					}
				}
			}
			return true
		})
		return true
	})
	return found
}

// catalogueEntry reports whether expr is a bare string literal that the
// catalogue would translate. A literal already wrapped in i18n.T is not a bare
// literal and never reaches here.
func catalogueEntry(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	text, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", false
	}
	if _, listed := i18n.Catalogue()[text]; !listed {
		return "", false
	}
	return text, true
}
