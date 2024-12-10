package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"testing"

	"user-service/internal/library/levels"
)

func Test(t *testing.T) {
	ctx := context.Background()
	slog.SetLogLoggerLevel(levels.Trace)
	t.Skip("Use Testing Tags ...")

	t.Run("New", func(t *testing.T) {
		t.Run("Default-Port", func(t *testing.T) {
			instance := New()

			if instance == nil {
				t.Fatalf("Instance is nil")
			} else if instance.options.Port != 8080 {
				t.Fatalf("Instance options port should be 8080, is %d", instance.options.Port)
			}
		})
	})

	t.Run("Start", func(t *testing.T) {
		instance := New()

		instance.Start(ctx)
		defer t.Cleanup(func() {
			instance.Stop(ctx)
		})

		if instance.Process() == nil {
			t.Fatalf("Instance Command Process is nil")
		}

		client := http.Client{}
		response, e := client.Get(fmt.Sprintf("http://localhost:%s/api/v1/namespaces", strconv.Itoa(instance.options.Port)))
		if e != nil {
			t.Fatalf("Unable to Call Proxied Namespaces API Endpoint: %v", e)
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatalf("Proxied Namespaces API Endpoint Returned Status Code: %d", response.StatusCode)
		}

		content, e := io.ReadAll(response.Body)
		if e != nil {
			t.Fatalf("Unable to Read Proxied Namespaces API Response Body: %v", e)
		}

		var structure map[string]interface{}
		if e := json.Unmarshal(content, &structure); e != nil {
			t.Fatalf("Unable to Unmarshal Proxied Namespaces API Response Body: %v", e)
		}

		t.Logf("Successfully Unmarshaled Proxied Namespaces API Response Body: %+v", structure)
	})
}
