package middleware

import "context"

type Valuer[T interface{}] interface {
	Value(ctx context.Context) T
}
