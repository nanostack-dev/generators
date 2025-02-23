package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type StructField struct {
	Name string
	Type string
	Tags map[string]string
}

type StructDef struct {
	Name       string
	Fields     []StructField
	PackageStr string
}

func ParseFile(filename string, typeName string) (*StructDef, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structDef StructDef
	structDef.PackageStr = node.Name.Name

	ast.Inspect(
		node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if s, ok := x.Type.(*ast.StructType); ok {
					if typeName != "" && x.Name.Name != typeName {
						return true
					}
					structDef.Name = x.Name.Name
					for _, field := range s.Fields.List {
						if len(field.Names) > 0 {
							structField := StructField{
								Name: field.Names[0].Name,
								Type: typeToString(field.Type),
								Tags: parseTags(field.Tag),
							}
							structDef.Fields = append(structDef.Fields, structField)
						}
					}
					return false
				}
			}
			return true
		},
	)

	return &structDef, nil
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", typeToString(t.X), t.Sel.Name)
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func parseTags(tag *ast.BasicLit) map[string]string {
	tags := make(map[string]string)
	if tag == nil {
		return tags
	}
	return tags
}
