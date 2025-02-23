package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanostack-dev/generators/builder/internal/generator"
	genparser "github.com/nanostack-dev/generators/builder/internal/parser"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", ".", "directory to scan for builder annotations")
	flag.Parse()

	if err := generateBuilders(dir); err != nil {
		log.Fatal(err)
	}
}

func generateBuilders(dir string) error {
	fset := token.NewFileSet()

	// Walk the directory and process each .go file
	return filepath.Walk(
		dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(
				path, "_test.go",
			) {
				return nil
			}

			// Parse the file
			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("parsing file %s: %w", path, err)
			}

			// Get package name
			pkgName := f.Name.Name

			// Look for structs with @builder annotation
			ast.Inspect(
				f, func(n ast.Node) bool {
					if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Doc != nil {
						if genDecl.Tok != token.TYPE {
							return true
						}

						for _, spec := range genDecl.Specs {
							typeSpec, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}

							structType, ok := typeSpec.Type.(*ast.StructType)
							if !ok {
								continue
							}

							// Check for @builder annotation
							hasBuilder := false
							for _, comment := range genDecl.Doc.List {
								if strings.Contains(comment.Text, "@builder") {
									hasBuilder = true
									break
								}
							}

							if hasBuilder {
								// Convert to our internal struct representation
								structDef := &genparser.StructDef{
									Name:       typeSpec.Name.Name,
									PackageStr: pkgName,
									Fields:     extractFields(structType),
								}

								// Generate builder
								outputFile := filepath.Join(
									filepath.Dir(path),
									strings.ToLower(structDef.Name)+"_builder.go",
								)
								if err := generator.Generate(
									structDef, pkgName, outputFile,
								); err != nil {
									log.Printf(
										"Error generating builder for %s: %v", structDef.Name, err,
									)
								}
							}
						}
					}
					return true
				},
			)

			return nil
		},
	)
}

func extractFields(structType *ast.StructType) []genparser.StructField {
	var fields []genparser.StructField
	for _, field := range structType.Fields.List {
		if len(field.Names) > 0 {
			fields = append(
				fields, genparser.StructField{
					Name: field.Names[0].Name,
					Type: typeToString(field.Type),
				},
			)
		}
	}
	return fields
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
