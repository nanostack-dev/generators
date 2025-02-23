package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	inputFile := filepath.Join("..", "..", "testdata", "person.go")
	outputFile := filepath.Join(tmpDir, "person_builder.go")

	// Set command line arguments
	os.Args = []string{
		"generator",
		"-file", inputFile,
		"-output", outputFile,
		"-package", "testmodel",
		"-type", "Person",
	}

	// Run main
	main()

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}
}
