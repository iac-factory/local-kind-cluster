package timeout_test

import (
	"math/rand"
	"net/http"
	"time"

	"verification-service/internal/library/middleware"
	"verification-service/internal/library/middleware/timeout"
)

func Example() {
	middlewares := middleware.Middleware()

	middlewares.Add(middleware.New().Timeout().Configuration(func(options *timeout.Settings) { options.Timeout = time.Second * 5 }).Middleware)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		process := time.Duration(rand.Intn(5)) * time.Second

		select {
		case <-ctx.Done():
			return

		case <-time.After(process):
			// The above channel simulates some hard work.
		}

		w.Write([]byte("done"))
	})

	http.ListenAndServe(":8080", middlewares.Handler(mux))
}
