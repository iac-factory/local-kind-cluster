package registration

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/api/avatar/types/update"
	"user-service/internal/database"
	"user-service/internal/library/middleware"
	"user-service/models/users"

	"user-service/internal/library/server"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "registration"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	// --> verify input

	var input Body
	if validator, e := server.Validate(ctx, update.V, r.Body, &input); e != nil {
		slog.WarnContext(ctx, "Unable to Verify Request Body")

		if validator != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(validator)

			return
		}

		http.Error(w, "Unable to Validate Request Body", http.StatusInternalServerError)
		return
	}

	slog.DebugContext(ctx, "Input", slog.Any("request", input))

	// --> establish database connection + transaction

	connection, e := database.Connection(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Connection to Database", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx, e := connection.Begin(ctx)
	if e != nil {
		slog.ErrorContext(ctx, "Error Establishing Database Transaction", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer database.Disconnect(ctx, connection, tx)

	// --> check for user existence

	count, e := users.New().Count(ctx, tx, input.Email)
	if e != nil {
		slog.ErrorContext(ctx, "Unable to Check if User Exists", slog.String("error", e.Error()))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if count >= 1 {
		const message = "Account With Email Address Already Exists"

		slog.ErrorContext(ctx, message, slog.String("email", input.Email))

		http.Error(w, message, http.StatusConflict)
		return
	}

	// --> create the user
	result, e := users.New().Create(ctx, tx, input.Email)
	if e != nil {
		const message = "Unable to Create New User"

		slog.ErrorContext(ctx, message, slog.String("email", input.Email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusConflict)
		return
	}

	// // --> establish stripe customer identifier channel to async wait for creation
	// identifier := make(chan string)
	// // --> establish optional for billing-service
	// proceed := make(chan struct{})

	// --> establish an errgroup for concurrency and fatal internal server error evaluation(s)

	// headers := telemetrics.New().Value(ctx).Headers
	//
	// c := telemetry.Client(headers)
	//
	// g := new(errgroup.Group)
	// g.Go(func() error { // --> notify and establish the user with verification-service
	// 	const service = "verification-service"
	//
	// 	namespace := os.Getenv("NAMESPACE")
	// 	if namespace == "" {
	// 		namespace = "development"
	// 	}
	//
	// 	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080/register", service, namespace)
	//
	// 	request, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	// 	if e != nil {
	// 		slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))
	//
	// 		return e
	// 	}
	//
	// 	response, e := c.Do(request)
	// 	if e != nil {
	// 		switch {
	// 		case strings.Contains(e.Error(), "no such host"):
	// 			proceed <- struct{}{}
	//
	// 			slog.WarnContext(ctx, "Verification-Service Registration Endpoint Not Found", slog.String("error", e.Error()))
	// 			// --> occurs during local testing due to lack of internal kubernetes networking
	// 			return nil
	// 		default:
	// 			slog.WarnContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))
	//
	// 			return e
	// 		}
	// 	}
	//
	// 	defer response.Body.Close()
	//
	// 	content, e := io.ReadAll(response.Body)
	// 	if e != nil {
	// 		slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))
	//
	// 		return e
	// 	}
	//
	// 	// --> only successful responses will be in json format
	//
	// 	switch response.StatusCode {
	// 	case http.StatusOK, http.StatusCreated:
	// 		var mapping map[string]interface{}
	// 		if e := json.Unmarshal(content, &mapping); e != nil {
	// 			slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("error", e.Error()))
	//
	// 			return e
	// 		}
	//
	// 		slog.InfoContext(ctx, "Service Returned a Successful Response", slog.String("service", service), slog.Any("value", mapping))
	// 	case http.StatusConflict:
	// 		slog.WarnContext(ctx, "Verification Record Already Exists in Verification-Service Database", slog.Bool("continue", true), slog.String("value", string(content)))
	// 	default:
	// 		slog.ErrorContext(ctx, "Service Returned a Fatal Error", slog.String("url", url), slog.Int("status", response.StatusCode), slog.String("response", string(content)))
	//
	// 		e = fmt.Errorf("service (%s) returned a fatal error (%d): %s", url, response.StatusCode, response.Status)
	//
	// 		return e
	// 	}
	//
	// 	return nil
	// })
	//
	// g.Go(func() error { // --> register the user with customer-service (stripe)
	// 	const service = "customer-service"
	//
	// 	namespace := os.Getenv("NAMESPACE")
	// 	if namespace == "" {
	// 		namespace = "development"
	// 	}
	//
	// 	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080/register", service, namespace)
	//
	// 	// var reader bytes.Buffer
	// 	// if e := json.NewEncoder(&reader).Encode(map[string]string{"email": input.Email}); e != nil {
	// 	// 	e = fmt.Errorf("unable to encode email address: %w", e)
	// 	//
	// 	// 	slog.ErrorContext(ctx, "Unable to Encode Email", slog.String("error", e.Error()))
	// 	//
	// 	// 	return e
	// 	// }
	//
	// 	request, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	// 	if e != nil {
	// 		slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("error", e.Error()))
	//
	// 		return e
	// 	}
	//
	// 	response, e := c.Do(request)
	// 	if e != nil {
	// 		switch {
	// 		case strings.Contains(e.Error(), "no such host"):
	// 			proceed <- struct{}{}
	//
	// 			slog.WarnContext(ctx, "Customer-Service Registration Endpoint Not Found", slog.String("error", e.Error()))
	// 			// --> occurs during local testing due to lack of internal kubernetes networking
	// 			return nil
	// 		default:
	// 			slog.WarnContext(ctx, "Unable to Send Request", slog.String("error", e.Error()))
	//
	// 			return e
	// 		}
	// 	}
	//
	// 	defer response.Body.Close()
	//
	// 	content, e := io.ReadAll(response.Body)
	// 	if e != nil {
	// 		slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("error", e.Error()))
	//
	// 		return e
	// 	}
	//
	// 	// --> only successful responses will be in json format
	//
	// 	switch response.StatusCode {
	// 	case http.StatusOK, http.StatusCreated:
	// 		var mapping map[string]interface{}
	// 		if e := json.Unmarshal(content, &mapping); e != nil {
	// 			slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("error", e.Error()))
	//
	// 			return e
	// 		}
	//
	// 		slog.InfoContext(ctx, "Service Returned a Successful Response", slog.String("service", service), slog.Any("value", mapping))
	//
	// 		v, ok := mapping["id"].(string)
	// 		if !(ok) {
	// 			slog.ErrorContext(ctx, "Unable to Retrieve Stripe Customer ID from Customer-Service")
	//
	// 			return errors.New("invalid typecast - unable to cast mapping[\"id\" to string")
	// 		}
	//
	// 		identifier <- v
	// 	default:
	// 		slog.ErrorContext(ctx, "Service Returned a Fatal Error", slog.String("url", url), slog.Int("status", response.StatusCode), slog.String("response", string(content)))
	//
	// 		e = fmt.Errorf("service (%s) returned a fatal error (%d): %s", url, response.StatusCode, response.Status)
	//
	// 		return e
	// 	}
	//
	// 	return nil
	// })
	//
	// g.Go(func() error { // stripe meters (billing-service)
	// 	const service = "billing-service"
	//
	// 	namespace := os.Getenv("NAMESPACE")
	// 	if namespace == "" {
	// 		namespace = "development"
	// 	}
	//
	// 	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:8080/meters", service, namespace)
	//
	// 	select {
	// 	case <-ctx.Done():
	// 		return ctx.Err()
	// 	case <-proceed:
	// 		return nil
	// 	case id := <-identifier:
	// 		// --> once id is hydrated, spawn two additional go-routines that will create two
	// 		// ... stripe meters: server-usage and storage-usage.
	//
	// 		server := Meter{
	// 			ID:       id,
	// 			Display:  "Server Usage",
	// 			Event:    "server-usage",
	// 			Key:      "minutes",
	// 			Type:     "server",
	// 			Metadata: map[string]string{},
	// 		}
	//
	// 		storage := Meter{
	// 			ID:       id,
	// 			Display:  "Storage Usage",
	// 			Event:    "storage-usage",
	// 			Key:      "minutes",
	// 			Type:     "storage",
	// 			Metadata: map[string]string{},
	// 		}
	//
	// 		child := new(errgroup.Group)
	// 		child.Go(func() error { // server meter
	// 			const context = "server"
	//
	// 			var reader bytes.Buffer
	// 			if e := json.NewEncoder(&reader).Encode(server); e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Encode Server Meter", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return fmt.Errorf("unable to encode server meter: %w", e)
	// 			}
	//
	// 			request, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			response, e := c.Do(request)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Send Request", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			defer response.Body.Close()
	//
	// 			content, e := io.ReadAll(response.Body)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			// --> only successful responses will be in json format
	//
	// 			switch response.StatusCode {
	// 			case http.StatusOK, http.StatusCreated:
	// 				var mapping map[string]interface{}
	// 				if e := json.Unmarshal(content, &mapping); e != nil {
	// 					slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 					return e
	// 				}
	//
	// 				slog.InfoContext(ctx, "Service Returned a Successful Response", slog.String("context", context), slog.String("service", service), slog.Any("value", mapping))
	// 			default:
	// 				slog.ErrorContext(ctx, "Service Returned a Fatal Error", slog.String("context", context), slog.String("url", url), slog.Int("status", response.StatusCode), slog.String("response", string(content)))
	//
	// 				e = fmt.Errorf("service (%s) returned a fatal error (%d): %s", url, response.StatusCode, response.Status)
	//
	// 				return e
	// 			}
	//
	// 			return nil
	// 		})
	//
	// 		child.Go(func() error { // storage meter
	// 			const context = "storage"
	//
	// 			var reader bytes.Buffer
	// 			if e := json.NewEncoder(&reader).Encode(storage); e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Encode Storage Meter", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return fmt.Errorf("unable to encode storage meter: %w", e)
	// 			}
	//
	// 			request, e := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Generate Request", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			response, e := c.Do(request)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Send Request", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			defer response.Body.Close()
	//
	// 			content, e := io.ReadAll(response.Body)
	// 			if e != nil {
	// 				slog.ErrorContext(ctx, "Unable to Read Raw Response", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 				return e
	// 			}
	//
	// 			// --> only successful responses will be in json format
	//
	// 			switch response.StatusCode {
	// 			case http.StatusOK, http.StatusCreated:
	// 				var mapping map[string]interface{}
	// 				if e := json.Unmarshal(content, &mapping); e != nil {
	// 					slog.ErrorContext(ctx, "Unable to Unmarshal Response", slog.String("context", context), slog.String("error", e.Error()))
	//
	// 					return e
	// 				}
	//
	// 				slog.InfoContext(ctx, "Service Returned a Successful Response", slog.String("context", context), slog.String("service", service), slog.Any("value", mapping))
	// 			default:
	// 				slog.ErrorContext(ctx, "Service Returned a Fatal Error", slog.String("context", context), slog.String("url", url), slog.Int("status", response.StatusCode), slog.String("response", string(content)))
	//
	// 				e = fmt.Errorf("service (%s) returned a fatal error (%d): %s", url, response.StatusCode, response.Status)
	//
	// 				return e
	// 			}
	//
	// 			return nil
	// 		})
	//
	// 		return g.Wait()
	// 	}
	// })
	//
	// if e := g.Wait(); e != nil { // an error is only returned upon a fatal, internal server error
	// 	slog.ErrorContext(ctx, "A Fatal Internal Server Error Has Occurred", slog.String("error", e.Error()))
	//
	// 	labeler.Add(attribute.Bool("error", true))
	// 	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	// 	return
	// }

	// --> commit the transaction only after all error cases have been evaluated

	if e := tx.Commit(ctx); e != nil {
		const message = "Unable to Commit Transaction"

		slog.ErrorContext(ctx, message, slog.String("email", input.Email))

		labeler.Add(attribute.Bool("error", true))
		http.Error(w, message, http.StatusConflict)
		return
	}

	slog.DebugContext(ctx, "Successfully Committed Database Transaction")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
