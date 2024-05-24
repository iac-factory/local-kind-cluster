package middleware

type Variadic[T interface{}] func(options *T)
