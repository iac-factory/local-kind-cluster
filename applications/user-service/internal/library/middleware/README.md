# `middleware` - HTTP Middleware

## Documentation

Official `godoc` documentation (with examples) can be found at the [Package Registry](https://pkg.go.dev/user-service/internal/library/middleware).

## Usage

###### Add Package Dependency

```bash
go get -u user-service/internal/library/middleware
```

###### Import & Implement

`main.go`

```go
package main

import (
    "encoding/json"
    "net/http"

    "user-service/internal/library/middleware"
    "user-service/internal/library/middleware/name"
)

func main() {
    middlewares := middleware.Middleware()
    middlewares.Add(middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = "example-service-name" }).Middleware)

    mux := http.NewServeMux()

    handler := middlewares.Handler(mux)

    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        value := middleware.New().Service().Value(ctx)

        var response = map[string]interface{}{
            "value": value,
        }

        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })

    http.ListenAndServe(":8080", handler)
}
```

- Please refer to the [code examples](example_test.go) for additional usage and implementation details.
- See https://pkg.go.dev/user-service/internal/library/middleware for additional documentation.

## Contributions

See the [**Contributing Guide**](CONTRIBUTING.md) for additional details on getting started.
