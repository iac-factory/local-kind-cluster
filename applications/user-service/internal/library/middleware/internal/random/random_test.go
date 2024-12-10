package random

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	const length = 8

	type Case struct {
		name        string
		parameters  struct{ length int }
		expectation int
	}

	tests := make([]Case, 1000)
	for index := range tests {
		(tests)[index] = Case{
			name: fmt.Sprintf("Random-Generator-%d", index),
			parameters: struct {
				length int
			}{
				length: length,
			},
			expectation: length,
		}
	}

	t.Run("Random-Test(s)", func(t *testing.T) {
		t.Parallel()
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				str := Random(tt.parameters.length)
				actual := len(str)
				expectation := tt.expectation

				if actual != expectation {
					t.Errorf("Actual Length Isn't Equal to User-Specification: (%d) != (%d)", actual, expectation)
				}

				fmt.Println(str)
			})
		}
	})
}
