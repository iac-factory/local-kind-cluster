package color

import (
	"testing"
)

func Test(t *testing.T) {
	Force()

	t.Run("Colors", func(t *testing.T) {
		Color().White("White").Black("Black").Cyan("Cyan").Red("Red").Yellow("Yellow").Green("Green").Gray("Gray").Purple("Purple").Blue("Blue").Default("Default").Write()
	})

	t.Run("CI", func(t *testing.T) {
		// override for uncolored output
		Force()

		t.Run("Color", func(t *testing.T) {
			Color().White("White").Black("Black").Cyan("Cyan").Red("Red").Yellow("Yellow").Green("Green").Gray("Gray").Purple("Purple").Blue("Blue").Default("Default").Write()
		})
	})

	t.Run("Bold", func(t *testing.T) {
		Color().Bold("Bold").Write()

		t.Run("Red", func(t *testing.T) {
			Color().Bold(Color().Red("Bold & Red").String()).Write()
		})

		t.Run("Green", func(t *testing.T) {
			Color().Bold(Color().Green("Bold & Green").String()).Write()
		})

		t.Run("Blue", func(t *testing.T) {
			Color().Bold(Color().Blue("Bold & Blue").String()).Write()
		})
	})

	t.Run("Italic", func(t *testing.T) {
		Color().Italic("Italic").Write()

		t.Run("Red", func(t *testing.T) {
			Color().Italic(Color().Red("Italic & Red").String()).Write()
		})

		t.Run("Green", func(t *testing.T) {
			Color().Italic(Color().Green("Italic & Green").String()).Write()
		})

		t.Run("Blue", func(t *testing.T) {
			Color().Italic(Color().Blue("Italic & Blue").String()).Write()
		})
	})

	t.Run("Special", func(t *testing.T) {
		Color().Bold(Color().Italic("Bold & Italic").String()).Write()

		t.Run("Red", func(t *testing.T) {
			Color().Bold(Color().Italic(Color().Red("Bold & Italic & Red").String()).String()).Write()
		})

		t.Run("Green", func(t *testing.T) {
			Color().Bold(Color().Italic(Color().Green("Bold & Italic & Green").String()).String()).Write()
		})

		t.Run("Blue", func(t *testing.T) {
			Color().Bold(Color().Italic(Color().Blue("Bold & Italic & Blue").String()).String()).Write()
		})
	})
}
