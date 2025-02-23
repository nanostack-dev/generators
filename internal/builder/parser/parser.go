package parser

import (
"fmt"
"go/ast"
"go/parser"
"go/token"
"strings"
)

type MethodMap struct {
From string
To   string
}

type StructField struct {
Name      string
Type      string
Tags      map[string]string
CustomGen bool // true if this field's setter should not be generated
}

type BuilderAnnotations struct {
Prefix      string      // @builder:prefix <value>
Validate    bool        // @builder:validate
Skip        bool        // @builder:skip
Package     string      // @builder:package <value>
Output      string      // @builder:output <pattern>
Immutable   bool        // @builder:immutable - generates Copy() instead of setters
Chain       bool        // @builder:chain - enables method chaining (default true)
Constructor string      // @builder:constructor <name> - custom constructor name
MethodMaps  []MethodMap // @builder:map <from>:<to> - maps one method to another
CustomMethods []string  // @builder:custom <method> - skip generation for these methods
}

type StructDef struct {
Name        string
Fields      []StructField
PackageStr  string
Imports     []string
Annotations BuilderAnnotations
}

// normalizeMethodName ensures consistent method name format for comparison
func normalizeMethodName(name string, prefix string) string {
name = strings.ToLower(name)
prefix = strings.ToLower(prefix)
if strings.HasPrefix(name, strings.ToLower(prefix)) {
return name
}
return prefix + name
}

// isCustomMethod checks if a method should be custom implemented
func isCustomMethod(fieldName string, prefix string, customMethods []string) bool {
normalizedFieldMethod := normalizeMethodName(fieldName, prefix)

for _, custom := range customMethods {
normalizedCustom := normalizeMethodName(custom, prefix)
if normalizedFieldMethod == normalizedCustom {
return true
}
}

return false
}

func ParseFile(filename string, typeName string) (*StructDef, error) {
fset := token.NewFileSet()
node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
if err != nil {
return nil, err
}

var structDef StructDef
structDef.PackageStr = node.Name.Name

// Collect imports
for _, imp := range node.Imports {
structDef.Imports = append(structDef.Imports, imp.Path.Value)
}

ast.Inspect(
node, func(n ast.Node) bool {
switch x := n.(type) {
case *ast.TypeSpec:
if s, ok := x.Type.(*ast.StructType); ok {
if typeName != "" && x.Name.Name != typeName {
return true
}
structDef.Name = x.Name.Name

// Parse struct annotations from doc comments
if x.Doc != nil {
structDef.Annotations = ParseAnnotations(x.Doc)
}

// Ensure we have a prefix
if structDef.Annotations.Prefix == "" {
structDef.Annotations.Prefix = "With"
}

// Process fields
for _, field := range s.Fields.List {
if len(field.Names) > 0 {
fieldName := field.Names[0].Name
structField := StructField{
Name: fieldName,
Type: typeToString(field.Type),
Tags: parseTags(field.Tag),
CustomGen: isCustomMethod(fieldName, structDef.Annotations.Prefix, structDef.Annotations.CustomMethods),
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
case *ast.SelectorExpr:
return fmt.Sprintf("%s.%s", typeToString(t.X), t.Sel.Name)
case *ast.ArrayType:
return "[]" + typeToString(t.Elt)
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

// ParseAnnotations extracts builder annotations from doc comments
func ParseAnnotations(comments *ast.CommentGroup) BuilderAnnotations {
annotations := BuilderAnnotations{
Prefix: "With",
Chain:  true,
}

if comments == nil {
return annotations
}

for _, comment := range comments.List {
text := strings.TrimPrefix(comment.Text, "//")
text = strings.TrimSpace(text)

switch {
case strings.HasPrefix(text, "@builder:prefix"):
value := strings.TrimPrefix(text, "@builder:prefix")
annotations.Prefix = strings.TrimSpace(value)
case strings.HasPrefix(text, "@builder:validate"):
annotations.Validate = true
case strings.HasPrefix(text, "@builder:skip"):
annotations.Skip = true
case strings.HasPrefix(text, "@builder:package"):
value := strings.TrimPrefix(text, "@builder:package")
annotations.Package = strings.TrimSpace(value)
case strings.HasPrefix(text, "@builder:output"):
value := strings.TrimPrefix(text, "@builder:output")
annotations.Output = strings.TrimSpace(value)
case strings.HasPrefix(text, "@builder:immutable"):
annotations.Immutable = true
case strings.HasPrefix(text, "@builder:nochain"):
annotations.Chain = false
case strings.HasPrefix(text, "@builder:constructor"):
value := strings.TrimPrefix(text, "@builder:constructor")
annotations.Constructor = strings.TrimSpace(value)
case strings.HasPrefix(text, "@builder:map"):
value := strings.TrimPrefix(text, "@builder:map")
value = strings.TrimSpace(value)
parts := strings.Split(value, ":")
if len(parts) == 2 {
annotations.MethodMaps = append(annotations.MethodMaps, MethodMap{
From: strings.TrimSpace(parts[0]),
To:   strings.TrimSpace(parts[1]),
})
}
case strings.HasPrefix(text, "@builder:custom"):
value := strings.TrimPrefix(text, "@builder:custom")
method := strings.TrimSpace(value)
annotations.CustomMethods = append(annotations.CustomMethods, method)
case text == "@builder":
// Base annotation, already handled
}
}

return annotations
}
