package testmodel

//go:generate go run github.com/nanostack-dev/generators/builder@latest gen -type Person -output person_builder.go

// @builder:prefix Set
// @builder:validate
type Person struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=0,lte=150"`
}
