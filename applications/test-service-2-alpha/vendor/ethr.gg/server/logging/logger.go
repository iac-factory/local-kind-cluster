package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"ethr.gg/server/logging/color"
)

// Type represents a string enumeration that defines the type of logger to use.
type Type string

const (
	// JSON represents a JSON SLOG Handler (slog.JSONHandler).
	JSON Type = "json"

	// Text represents a Text SLOG Handler (slog.TextHandler).
	Text Type = "text"
)

// logger represents an atomic pointer to enumeration Type - must be one of ( JSON | Text | Default ).
var logger atomic.Value // atomic.Pointer[Type]

// Global will update the atomic logger value of which the slog.Handler is constructed.
func Global(v Type) {
	switch v {
	case JSON:
		logger.Store(JSON)
	case Text:
		logger.Store(Text)
	default:
		logger.Store(Text)
	}
}

// Default represents the value derived from the atomic.Value variable defined at "logger".
//
//   - Please see Global for updating the pointer value.
func Default() Type {
	return logger.Load().(Type)
}

var verbose atomic.Pointer[bool]

// Verbose enables an atomic.Pointer to log internal Handler log messages.
func Verbose(v bool) {
	verbose.Store(&v)
}

func init() {
	var handler = Text
	if v := os.Getenv("LOGGER"); v != "" {
		switch v {
		case "json":
			handler = JSON
		case "text":
			handler = Text
		default:
			log.Println("WARNING: invalid LOGGER environment variable - must be \"json\" or \"text\". Defaulting to \"text\"")
		}
	}

	logger.Store(handler)

	Verbose(false)
	if v := strings.ToLower(os.Getenv("VERBOSE")); v == "enabled" || v == "true" || v == "1" {
		Verbose(true)
	}
}

type Options struct {
	Service  string
	Settings *slog.HandlerOptions
}

type Handler struct {
	slog.Handler

	service  string
	settings *slog.HandlerOptions

	writer io.Writer

	t    Type
	text *slog.TextHandler
	json *slog.JSONHandler

	logger *log.Logger
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	h.logger.SetOutput(h.writer)

	minimum := slog.LevelInfo
	if h.settings == nil {
		h.settings = &slog.HandlerOptions{}
	} else if h.settings.Level != nil {
		minimum = h.settings.Level.Level()
	}

	enabled := level >= minimum
	if verbose.Load() != nil && *(verbose.Load()) {
		fmt.Fprintf(os.Stdout, "Evaluating Logger Enablement - Instance Log Level: %s, Function Log Level: %s, Enabled: %v \n", minimum, level, enabled)
	}

	return enabled
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	if h.t == JSON {
		h.json.Handle(ctx, record)
	}

	var level string

	switch record.Level {
	case Trace:
		level = fmt.Sprintf("%s:", color.Color().Dim("TRACE").String())
	case Debug:
		level = fmt.Sprintf("%s:", color.Color().Purple("DEBUG").String())
	case Info:
		level = fmt.Sprintf("%s:", color.Color().Green("INFO").String())
	case Warn:
		level = fmt.Sprintf("%s:", color.Color().Yellow("WARN").String())
	case Error:
		level = fmt.Sprintf("%s:", color.Color().Red("ERROR").String())
	default:
		return fmt.Errorf("invalid, unknown level: %s", record.Level.String())
	}

	fields := make(map[string]interface{}, record.NumAttrs())
	var evaluate = func(a slog.Attr) bool {
		if verbose.Load() != nil && *(verbose.Load()) {
			fmt.Fprintf(os.Stdout, "%s: %s\n", color.Color().Cyan(a.Key).String(), a.String())
		}

		if kind := a.Value.Kind(); kind == slog.KindGroup {
			group := a.Value.Group()
			mapping := make(map[string]interface{}, len(group))

			for index := range group {
				attribute := group[index]
				value := attribute.Value.Any()
				if kind := attribute.Value.Kind(); kind == slog.KindGroup {
					child := attribute.Value.Group()
					children := make(map[string]interface{}, len(child))
					for index := range attribute.Value.Group() {
						sibling := child[index]
						assignment := sibling.Value.Any()
						if kind := sibling.Value.Kind(); kind == slog.KindGroup {
							final := sibling.Value.Group()
							nesting := make(map[string]interface{}, len(final))
							for index := range final {
								nest := final[index]
								v := nest.Value.Any()

								if kind := nest.Value.Kind(); kind == slog.KindGroup {
									nesting[sibling.Key] = nest.Value.String()
								} else {
									nesting[sibling.Key] = v
								}
							}

							children[sibling.Key] = nesting
						} else {
							children[sibling.Key] = assignment
						}
					}

					mapping[attribute.Key] = children
				} else {
					mapping[attribute.Key] = value
				}
			}

			fields[a.Key] = mapping

			return true
		}

		value := a.Value.Any()
		switch value.(type) {
		case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
			fields[a.Key] = value
		default:
			output, e := json.Marshal(a.Value.Any())
			if e != nil {
				fmt.Fprintf(os.Stderr, "ERROR - Unable to Marshal Logging Attribute: %s\n", a.Value.String())

				return false
			}

			if e := json.Unmarshal(output, &value); e != nil {
				fmt.Fprintf(os.Stderr, "ERROR - Unable to Unmarshal Logging Attribute: %s\n", a.Value.String())

				return false
			}

			fields[a.Key] = value
			if a.Key == "$" && fields[a.Key] != nil { // special key to denote the mapping should be inlined.
				if v, ok := value.(map[string]interface{}); ok {
					fields = v
				}
			}
		}

		return true
	}

	record.Attrs(evaluate)

	buffer, e := json.MarshalIndent(fields, "", "    ")
	if e != nil {
		e = fmt.Errorf("failed to marshal fields to json: %v", e)
		return e
	}

	format := record.Time.Format("[15:05:05.000]")
	message := color.Cyan(record.Message)

	if service := h.service; service != "" && h.logger.Prefix() == "" {
		literal := color.Color().Bold(color.Color().Red(service).String()).String()

		prefix := fmt.Sprintf("%s - ", literal)

		h.logger.SetPrefix(prefix)
	}

	h.logger.Println(format, level, message, color.White(string(buffer)))

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.Handler.WithAttrs(attrs)
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return h.Handler.WithGroup(name)
}

func Logger(w io.Writer, options Options) slog.Handler {
	if options.Settings == nil {
		options.Settings = &slog.HandlerOptions{}
	}

	var instantiation = &Handler{writer: w, service: options.Service, settings: options.Settings, logger: log.New(w, "", 0)}

	instantiation.t = Text
	instantiation.text = slog.NewTextHandler(w, options.Settings)
	instantiation.Handler = instantiation.text

	if Default() == JSON {
		instantiation.t = JSON
		instantiation.json = slog.NewJSONHandler(w, options.Settings)
		instantiation.Handler = instantiation.json
	}

	return instantiation
}
