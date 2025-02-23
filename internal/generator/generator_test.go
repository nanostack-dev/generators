package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nanostack/generator/internal/parser"
)

func TestGenerate(t *testing.T) {
	// Test struct definition
	structDef := &parser.StructDef{
		Name:       "Person",
		PackageStr: "testmodel",
		Fields: []parser.StructField{
			{Name: "Name", Type: "string"},
			{Name: "Age", Type: "int"},
		},
	}

	// Create temporary output file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "person_builder.go")

	// Generate builder
	err := Generate(structDef, "testmodel", outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read generated file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Basic validation
	generated := string(content)
	requiredElements := []string{
		"package testmodel",
		"type PersonBuilder struct",
		"func NewPersonBuilder() *PersonBuilder",
		"func (b *PersonBuilder) WithName(name string) *PersonBuilder",
		"func (b *PersonBuilder) WithAge(age int) *PersonBuilder",
		"func (b *PersonBuilder) Build() *Person",
	}

	t.Logf("Generated code:\n%s", generated)
	for _, element := range requiredElements {
		if !strings.Contains(generated, element) {
			t.Errorf("Generated code missing: %s", element)
		}
	}
}
