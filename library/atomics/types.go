package atomics

import (
	"log/slog"
	"sync/atomic"
)

type String struct {
	pointer atomic.Value
}

func (v String) Set(value string) {
	v.pointer.Store(value)
}

func (v String) Get() string {
	if v.pointer.Load() == nil {
		slog.Warn("Invalid Atomic Assignment - Returning Empty String")
		return ""
	}

	return v.pointer.Load().(string)
}

type Map struct {
	pointer atomic.Value
}

func (v Map) Set(key, value string) {
	var assignment map[string]string
	if v.pointer.Load() == nil {
		v.pointer.Store(make(map[string]string))
	}

	assignment = v.pointer.Load().(map[string]string)

	assignment[key] = value

	v.pointer.Store(assignment)
}

func (v Map) Get() map[string]string {
	if v.pointer.Load() == nil {
		v.pointer.Store(make(map[string]string))
	}

	return v.pointer.Load().(map[string]string)
}

//
// type Handler func(pattern string, function http.HandlerFunc) http.Handler
//
// type Route struct {
// 	index   int
// 	pattern string
// 	handler Handler
// }
//
// type Router struct {
// 	pointer atomic.Value
// }
//
// func (v Router) Set(pattern string, handlers ...Handler) {
// 	var assignment map[string]Route
// 	if v.pointer.Load() == nil {
// 		v.pointer.Store(make(map[string]Route))
// 	}
//
// 	assignment = v.pointer.Load().(map[string]Route)
//
// 	for index, handler := range handlers {
// 		assignment[pattern] = Route{index: index, pattern: pattern, handler: handler}
// 	}
//
// 	assignment[pattern] = handler
//
// 	v.pointer.Store(assignment)
// }
//
// func (v Router) Get() map[string]Handler {
// 	if v.pointer.Load() == nil {
// 		v.pointer.Store(make(map[string]Handler))
// 	}
//
// 	return v.pointer.Load().(map[string]Handler)
// }
