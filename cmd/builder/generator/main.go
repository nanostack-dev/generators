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

"github.com/nanostack-dev/generators/internal/builder/generator"
genparser "github.com/nanostack-dev/generators/internal/builder/parser"
)

type config struct {
dir             string
prefix          string
outputPattern   string
packageOverride string
validate        bool
}

func main() {
// Define default configuration
defaultCfg := config{
dir:           ".",
prefix:        "With",
outputPattern: "{name}_builder.go",
}

// Create config to store flag values
cfg := config{}

// Setup flags with defaults
flag.StringVar(&cfg.dir, "dir", defaultCfg.dir, "directory to scan for builder annotations")
flag.StringVar(&cfg.prefix, "prefix", defaultCfg.prefix, "prefix for builder methods (default: With)")
flag.StringVar(&cfg.outputPattern, "output", defaultCfg.outputPattern, "output file pattern. Use {name} as placeholder for struct name")
flag.StringVar(&cfg.packageOverride, "package", "", "override package name")
flag.BoolVar(&cfg.validate, "validate", false, "generate validation methods")
flag.Parse()

log.Printf("Generating builders with config: %+v\n", cfg)

if err := generateBuilders(cfg, defaultCfg); err != nil {
log.Fatal(err)
}
}

func generateBuilders(cfg config, defaultCfg config) error {
fset := token.NewFileSet()

return filepath.Walk(cfg.dir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}

if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
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
ast.Inspect(f, func(n ast.Node) bool {
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

hasBuilder := false
for _, comment := range genDecl.Doc.List {
if strings.Contains(comment.Text, "@builder") {
hasBuilder = true
break
}
}

if hasBuilder {
// Parse annotations first
annotations := genparser.ParseAnnotations(genDecl.Doc)

// Create struct definition
structDef := &genparser.StructDef{
Name:       typeSpec.Name.Name,
PackageStr: pkgName,
Fields:     extractFields(structType),
// Initialize Annotations with defaults
Annotations: genparser.BuilderAnnotations{
Chain: true, // Default to true for method chaining
},
}

// Only apply CLI values if they differ from defaults
if flag.Lookup("prefix").Value.String() != defaultCfg.prefix {
structDef.Annotations.Prefix = cfg.prefix
} else {
structDef.Annotations.Prefix = annotations.Prefix
}

if flag.Lookup("validate").Value.String() == "true" {
structDef.Annotations.Validate = true
} else {
structDef.Annotations.Validate = annotations.Validate
}

// Copy other annotation values
structDef.Annotations.Skip = annotations.Skip
structDef.Annotations.Package = annotations.Package
structDef.Annotations.Output = annotations.Output
structDef.Annotations.Immutable = annotations.Immutable
structDef.Annotations.Chain = annotations.Chain // This will keep true unless explicitly set to false
structDef.Annotations.Constructor = annotations.Constructor

// Collect imports
for _, imp := range f.Imports {
structDef.Imports = append(structDef.Imports, imp.Path.Value)
}

// Determine output pattern
outputPattern := cfg.outputPattern
if annotations.Output != "" {
outputPattern = annotations.Output
}

// Generate output file name
outputName := strings.ReplaceAll(outputPattern, "{name}", strings.ToLower(structDef.Name))
outputFile := filepath.Join(filepath.Dir(path), outputName)

// Determine package name
pkgToUse := pkgName
if annotations.Package != "" {
pkgToUse = annotations.Package
} else if cfg.packageOverride != "" {
pkgToUse = cfg.packageOverride
}

if err := generator.Generate(structDef, pkgToUse, outputFile); err != nil {
log.Printf("Error generating builder for %s: %v", structDef.Name, err)
}
}
}
}
return true
})

return nil
})
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
