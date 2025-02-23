package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"nanostack/generator/internal/generator"
	"nanostack/generator/internal/parser"
)

func main() {
	inputFile := flag.String(
		"file", os.Getenv("GOFILE"), "Input file containing the struct definition",
	)
	outputFile := flag.String("output", "", "Output file for the generated builder")
	packageName := flag.String("package", "", "Package name for the generated builder")
	typeName := flag.String(
		"type", os.Getenv("GOTYPE"), "Name of the struct to generate builder for",
	)
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("Input file is required")
	}

	if *outputFile == "" {
		base := strings.TrimSuffix(filepath.Base(*inputFile), ".go")
		*outputFile = fmt.Sprintf("%s_builder.go", base)
	}

	structDef, err := parser.ParseFile(*inputFile, *typeName)
	if err != nil {
		log.Fatalf("Failed to parse input file: %v", err)
	}

	if err := generator.Generate(structDef, *packageName, *outputFile); err != nil {
		log.Fatalf("Failed to generate builder: %v", err)
	}
}
