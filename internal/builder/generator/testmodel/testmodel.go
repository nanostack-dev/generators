package testmodel

//go:generate go run github.com/nanostack-dev/generators/cmd/builder@latest .

// @builder
// @builder:prefix Set
type Person struct {
	Name string
	Age  int
}
