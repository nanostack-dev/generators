package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"nanostack/generator/internal/parser"
)

// Generate creates a builder pattern implementation for the given struct
func Generate(structDef *parser.StructDef, packageName string, outputFile string) error {
	if structDef == nil {
		return fmt.Errorf("structDef cannot be nil")
	}

	if packageName == "" {
		packageName = structDef.PackageStr
	}

	f := jen.NewFile(packageName)

	// Add generation comment
	f.HeaderComment("Code generated by nanostack/generator; DO NOT EDIT.")

	builderName := structDef.Name + "Builder"

	// Generate builder struct
	f.Type().Id(builderName).Struct(
		jen.Id("instance").Op("*").Id(structDef.Name),
	)

	// Generate constructor
	generateConstructor(f, builderName, structDef.Name)

	// Generate ToBuilder method
	generateToBuilder(f, builderName, structDef)

	// Generate With methods for each field
	for _, field := range structDef.Fields {
		generateWithMethod(f, builderName, field, structDef.Name)
	}

	// Generate Build method
	generateBuildMethod(f, builderName, structDef.Name)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return f.Save(outputFile)
}

func generateConstructor(f *jen.File, builderName, structName string) {
	f.Func().Id("New" + builderName).Params().Op("*").Id(builderName).Block(
		jen.Return(
			jen.Op("&").Id(builderName).Values(
				jen.Id("instance").Op(":").Op("&").Id(structName).Values(),
			),
		),
	)
}

func generateToBuilder(f *jen.File, builderName string, structDef *parser.StructDef) {
	f.Func().Params(
		jen.Id("p").Op("*").Id(structDef.Name),
	).Id("ToBuilder").Params().Op("*").Id(builderName).Block(
		jen.If(jen.Id("p").Op("==").Nil()).Block(
			jen.Return(jen.Id("New"+builderName).Call()),
		),
		jen.Return(
			jen.Op("&").Id(builderName).Values(
				jen.Id("instance").Op(":").Op("&").Id(structDef.Name).Values(
					generateFieldAssignments(structDef)...,
				),
			),
		),
	)
}

func generateBuildMethod(f *jen.File, builderName, structName string) {
	f.Func().Params(
		jen.Id("b").Op("*").Id(builderName),
	).Id("Build").Params().Op("*").Id(structName).Block(
		jen.Return(jen.Id("b").Dot("instance")),
	)
}

func generateWithMethod(
	f *jen.File, builderName string, field parser.StructField, structName string,
) {
	f.Func().Params(
		jen.Id("b").Op("*").Id(builderName),
	).Id("With"+field.Name).Params(
		jen.Id(strings.ToLower(field.Name)).Id(field.Type),
	).Op("*").Id(builderName).Block(
		jen.Id("b").Dot("instance").Dot(field.Name).Op("=").Id(strings.ToLower(field.Name)),
		jen.Return(jen.Id("b")),
	)
}

func generateFieldAssignments(structDef *parser.StructDef) []jen.Code {
	assignments := make([]jen.Code, len(structDef.Fields))
	for i, field := range structDef.Fields {
		assignments[i] = jen.Id(field.Name).Op(":").Id("p").Dot(field.Name)
	}
	return assignments
}
