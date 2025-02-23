package testmodel

//go:generate go run github.com/nanostack-dev/generators/cmd/builder/generator .

// @builder
// @builder:prefix Set
type Person struct {
	Name string
	Age  int
}
