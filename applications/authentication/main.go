package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"golang.org/x/crypto/bcrypt"

	"authentication/internal/name"
)

func dsn() string {
	host := os.Getenv("POSTGRES_SERVICE_HOST")
	port := os.Getenv("POSTGRES_SERVICE_PORT")

	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DB")

	mode := "prefer"
	timeout := "10"
	tz := "UTC"

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s application_name=%s sslmode=%s TimeZone=%s",
		host, port, username, password, database, timeout, name.Name, mode, tz,
	)
}

func open(ctx context.Context) (*pgx.Conn, error) {
	connection, e := pgx.Connect(ctx, dsn())
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Create Database Connection", slog.Group(name.Name, slog.String("error", e.Error())))
		return nil, e
	}

	if e := connection.Ping(ctx); e != nil {
		slog.ErrorContext(ctx, "Unable to Ping Database", slog.Group(name.Name, slog.String("error", e.Error())))
		return nil, e
	}

	return connection, nil
}

func disconnect(ctx context.Context, connection *pgx.Conn) error {
	if connection != nil {
		if e := connection.Close(ctx); e != nil {
			slog.ErrorContext(ctx, "Unable to Close Database Connection", slog.Group(name.Name, slog.String("error", e.Error())))
			return e
		}

		slog.DebugContext(ctx, "Successfully Closed Database Connection", slog.String("Application", name.Name))
	}

	return nil
}

func validate(ctx context.Context, hash, password string) bool {
	if e := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); e != nil {
		slog.WarnContext(ctx, "Invalid Hash-Password Comparison", slog.Group(name.Name, slog.String("error", e.Error())))

		return false
	}

	return true
}

func health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slog.DebugContext(ctx, "Health Check", slog.String("Application", name.Name))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "Healthy",
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connection, e := open(ctx)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		defer disconnect(ctx, connection)
		return
	}

	defer disconnect(ctx, connection)

	var input map[string]interface{}
	if e := json.NewDecoder(r.Body).Decode(&input); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	var valid bool

	var username string
	if value, ok := input["username"]; !(ok) {
		http.Error(w, "username field required", http.StatusBadRequest)
		return
	} else if username, valid = value.(string); !(valid) {
		http.Error(w, "username field must be a string", http.StatusBadRequest)
		return
	} else if len(username) == 0 {
		http.Error(w, "non empty username field (string) is required", http.StatusBadRequest)
		return
	}

	var password string
	if value, ok := input["password"]; !(ok) {
		http.Error(w, "password field required", http.StatusBadRequest)
		return
	} else if password, valid = value.(string); !(valid) {
		http.Error(w, "password field must be a string", http.StatusBadRequest)
		return
	} else if len(password) == 0 {
		http.Error(w, "non empty password field (string) is required", http.StatusBadRequest)
		return
	}

	type record struct {
		ID           int    `json:"id"`
		Username     string `json:"username"`
		Password     string `json:"-"`
		Verification string `json:"verification"`
	}

	var user = new(record)

	{
		rows, e := connection.Query(ctx, "SELECT \"id\", \"username\", \"password\", \"verification-status\" FROM \"User\" WHERE (\"deletion\" IS NULL) AND (\"username\" = $1);", username)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Query Database Row(s)", slog.String("error", e.Error()))
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		var found, validated bool
		for rows.Next() && !(found) {
			if e := rows.Err(); e != nil {
				slog.ErrorContext(ctx, "Unable to Iterate Database Row", slog.String("error", e.Error()))
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}

			if e := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Verification); e != nil {
				slog.ErrorContext(ctx, "Unable to Scan Database Row", slog.String("error", e.Error()))
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}

			found = true
		}

		validated = validate(ctx, user.Password, password)

		if !(found) || !(validated) {
			if !(found) {
				slog.WarnContext(ctx, "Unable to Find User", slog.String("username", username))
			} else if !(validated) {
				slog.WarnContext(ctx, "Incorrect Password", slog.String("username", username))
			}

			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	if user.Verification != "VERIFIED" {
		slog.WarnContext(ctx, "User Unverified", slog.String("username", username))
		http.Error(w, "user verification field must be \"VERIFIED\"", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	variables := os.Environ()
	for _, variable := range variables {
		partials := strings.Split(variable, "=")

		if len(partials) == 2 {
			slog.Debug("Environment Variable", slog.Group(name.Name, slog.String("Variable", partials[0]), slog.String("Value", partials[1])))
		}
	}

	// Handle requests at the root path using the handler function
	http.HandleFunc("POST /", handler)

	// Handle requests at the root path using the handler function
	http.HandleFunc("GET /health", health)

	http.ListenAndServe(":5000", nil)
}
