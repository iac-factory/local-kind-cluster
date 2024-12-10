package timeout

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test(t *testing.T) {
	tests := []struct {
		name       string
		middleware func(next http.Handler) http.Handler
		status     int
		handler    http.HandlerFunc
	}{
		{
			name:       "Successful-Response",
			middleware: New().Configuration(func(options *Settings) { options.Timeout = time.Second * 5 }).Middleware,
			status:     200,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				select {
				case <-ctx.Done():
					return

				case <-time.After(1 * time.Second):
					datum := struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					}{
						Key:   "Key",
						Value: "Value",
					}

					buffer, _ := json.MarshalIndent(datum, "", "    ")

					defer w.Write(buffer)

					w.Header().Set("Content-Type", "application/json")

					w.WriteHeader(http.StatusOK)

					return
				}
			}),
		},
		{
			name:       "Unsuccessful-Response",
			middleware: New().Configuration(func(options *Settings) { options.Timeout = time.Second * 5 }).Middleware,
			status:     504,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				select {
				case <-ctx.Done():
					return

				case <-time.After(30 * time.Second):
					datum := struct {
						Key   string `json:"key"`
						Value string `json:"value"`
					}{
						Key:   "Key",
						Value: "Value",
					}

					buffer, _ := json.MarshalIndent(datum, "", "    ")

					defer w.Write(buffer)

					w.Header().Set("Content-Type", "application/json")

					w.WriteHeader(http.StatusOK)

					return
				}
			}),
		},
	}

	for _, matrix := range tests {
		t.Run(matrix.name, func(t *testing.T) {
			server := httptest.NewServer(matrix.middleware(matrix.handler))

			defer server.Close()

			client := server.Client()
			request, e := http.NewRequest(http.MethodGet, server.URL, nil)
			if e != nil {
				t.Fatal(e)
			}

			response, exception := client.Do(request)
			if exception != nil {
				t.Fatal(exception)
			}

			defer response.Body.Close()

			status := response.StatusCode
			if status != matrix.status {
				t.Errorf("Status = %d\n    - Expectation = %d", status, matrix.status)
			}

			t.Logf("Successful\nStatus = %d\n    - Expectation = %d", status, matrix.status)
		})
	}

}
