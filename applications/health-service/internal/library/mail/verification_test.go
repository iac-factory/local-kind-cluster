package mail

import (
	"context"
	"testing"

	"health-service/internal/library/random"
)

func TestVerification(t *testing.T) {
	ctx := context.Background()

	t.Skip("Skipping Integration Tests")
	t.Run("Email", func(t *testing.T) {
		t.Run("Successful-Submission", func(t *testing.T) {
			if e := Verification(ctx, "jsanders4129@gmail.com", random.Verification()); e != nil {
				t.Errorf("Verification Returned non-nil Error: %s", e.Error())
			}
		})
	})
}
