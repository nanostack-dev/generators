// generator_test.go
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanostack-dev/generators/internal/builder/parser"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name          string
		structDef     *parser.StructDef
		packageName   string
		expectError   bool
		validateItems []string
	}{
		{
			name: "basic_struct",
			structDef: &parser.StructDef{
				Name:       "Person",
				PackageStr: "testmodel",
				Fields: []parser.StructField{
					{Name: "Name", Type: "string"},
					{Name: "Age", Type: "int"},
				},
			},
			packageName: "testmodel",
			validateItems: []string{
				"package testmodel",
				"type PersonBuilder struct",
				"func NewPersonBuilder() *PersonBuilder",
				"func (b *PersonBuilder) WithName(name string) *PersonBuilder",
				"func (b *PersonBuilder) WithAge(age int) *PersonBuilder",
				"func (b *PersonBuilder) Build() *Person",
			},
		},
		{
			name: "with_prefix",
			structDef: &parser.StructDef{
				Name:       "User",
				PackageStr: "testmodel",
				Fields: []parser.StructField{
					{Name: "Email", Type: "string"},
					{Name: "Active", Type: "bool"},
				},
				Annotations: parser.BuilderAnnotations{
					Prefix: "Set",
				},
			},
			packageName: "testmodel",
			validateItems: []string{
				"package testmodel",
				"type UserBuilder struct",
				"func NewUserBuilder() *UserBuilder",
				"func (b *UserBuilder) SetEmail(email string) *UserBuilder",
				"func (b *UserBuilder) SetActive(active bool) *UserBuilder",
				"func (b *UserBuilder) Build() *User",
			},
		},
		{
			name: "with_skip_annotation",
			structDef: &parser.StructDef{
				Name:       "Config",
				PackageStr: "testmodel",
				Annotations: parser.BuilderAnnotations{
					Skip: true,
				},
			},
			packageName:   "testmodel",
			validateItems: []string{},
		},
		{
			name:          "nil_struct_def",
			structDef:     nil,
			packageName:   "testmodel",
			expectError:   true,
			validateItems: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				outputFile := filepath.Join(tmpDir, strings.ToLower(tt.name)+"_builder.go")

				err := Generate(tt.structDef, tt.packageName, outputFile)

				if tt.expectError && err == nil {
					t.Error("expected error but got none")
					return
				}
				if !tt.expectError && err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if tt.expectError {
					return
				}

				if tt.structDef != nil && tt.structDef.Annotations.Skip {
					if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
						t.Error("file was generated despite skip annotation")
					}
					return
				}

				content, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				generated := string(content)
				for _, element := range tt.validateItems {
					if !strings.Contains(generated, element) {
						t.Errorf("Generated code missing: %s", element)
					}
				}

				if !strings.HasPrefix(generated, "// Code generated") {
					t.Error("Missing generation comment")
				}

				t.Logf("Generated code:\n%s", generated)
			},
		)
	}
}

func TestGenerateWithCustomTypes(t *testing.T) {
	structDef := &parser.StructDef{
		Name:       "Order",
		PackageStr: "testmodel",
		Fields: []parser.StructField{
			{Name: "ID", Type: "uuid.UUID"},
			{Name: "Items", Type: "[]Item"},
			{Name: "Status", Type: "OrderStatus"},
			{Name: "Created", Type: "time.Time"},
		},
	}

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "order_builder.go")

	err := Generate(structDef, "testmodel", outputFile)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	generated := string(content)
	expectedTypes := []string{
		"uuid.UUID",
		"[]Item",
		"OrderStatus",
		"time.Time",
	}

	for _, typ := range expectedTypes {
		if !strings.Contains(generated, typ) {
			t.Errorf("Generated code missing type: %s", typ)
		}
	}
}
