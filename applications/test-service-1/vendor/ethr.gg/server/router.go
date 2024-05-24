package server

// func Router(routes ...Handle) *http.ServeMux {
// 	mux := Mux()
//
// 	for _, route := range routes {
// 		var m *http.ServeMux
// 		var pattern string
// 		var function http.HandlerFunc
//
// 		instantiate := func(in []reflect.Value) []reflect.Value {
// 			m = in[0].Interface().(*http.ServeMux)
// 			pattern = in[1].Interface().(string)
// 			function = in[2].Interface().(http.HandlerFunc)
//
// 			return in
// 		}
//
// 		wrap := func(pointer interface{}) {
// 			// pointer is a pointer to a function.
//
// 			// Obtain the function value itself (likely nil) as a reflect.Value
// 			// so that we can query its type and then set the value.
// 			fn := reflect.ValueOf(pointer).Elem()
//
// 			// Make a function of the right type
// 			v := reflect.MakeFunc(fn.Type(), instantiate)
//
// 			// Assign it to the value fn represents.
// 			fn.Set(v)
// 		}
//
// 		wrap(&route)
//
// 		fmt.Println(pattern)
//
// 		_ = function
// 		_ = m
// 	}
//
// 	return mux
// }
