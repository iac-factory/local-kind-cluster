package mail

import (
	"bytes"
	"os"
	"testing"
)

func TestTemplate(t *testing.T) {
	t.Run("Templates", func(t *testing.T) {
		t.Run("Text", func(t *testing.T) {
			t.Run("Template", func(t *testing.T) {
				if v := Text.Template(); v != nil {
					t.Logf("Successfully Validated Text Template")
				} else {
					t.Errorf("Invalid Text Template")
				}
			})

			t.Run("Execution", func(t *testing.T) {
				var buffer bytes.Buffer

				e := Text.Execute(&buffer, Metadata{
					Expiration: 5,
					Duration:   "days",
					URL:        "https://testing.ethr.gg/verification",
				})

				if e != nil {
					t.Errorf("Unable to Render Text Template")
				}

				_ = os.WriteFile("email.txt", buffer.Bytes(), os.ModePerm)
			})
		})

		t.Run("HTML", func(t *testing.T) {
			t.Run("Template", func(t *testing.T) {
				if v := HTML.Template(); v != nil {
					t.Logf("Successfully Validated HTML Template")
				} else {
					t.Errorf("Invalid HTML Template")
				}
			})

			t.Run("Execution", func(t *testing.T) {
				var buffer bytes.Buffer

				e := HTML.Execute(&buffer, Metadata{
					Expiration: 5,
					Duration:   "days",
					URL:        "https://testing.ethr.gg/verification",
				})

				if e != nil {
					t.Errorf("Unable to Render HTML Template")
				}

				_ = os.WriteFile("index.html", buffer.Bytes(), os.ModePerm)
			})
		})
	})
}
