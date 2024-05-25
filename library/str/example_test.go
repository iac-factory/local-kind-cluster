package str_test

import (
	"fmt"

	"ethr.gg/str"
)

func ExampleLowercase() {
	v := "Field-Name"

	fmt.Println(str.Lowercase(v))
	// Output: field-name
}

func ExampleTitle() {
	v := "jacob b. sanders"

	fmt.Println(str.Title(v))
	// Output: Jacob B. Sanders
}

func ExampleVariadic() {
	v := str.Dereference(nil, func(o str.Options) {
		o.Log = true
	})

	fmt.Println(v)
	// Output:

	pointer := str.Pointer("example")
	v = str.Dereference(pointer, func(o str.Options) {
		o.Log = true
	})

	fmt.Println(v)
	// Output: example
}

func ExampleDereference() {
	pointer := str.Pointer("example")
	v := str.Dereference(pointer, func(o str.Options) {
		o.Log = true
	})

	fmt.Println(v)
	// Output: example
}

func ExamplePointer() {
	pointer := str.Pointer("example", func(o str.Options) {
		o.Log = true
	})

	fmt.Println(*(pointer))
	// Output: example
}
