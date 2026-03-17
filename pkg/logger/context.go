package logger

import "context"

type ctxKey string

type ctxValue map[string]interface{}

const (
	ContextLogKey ctxKey = "context_log"
)

func WithLogContext(ctx context.Context, key string, value interface{}) context.Context {
	if c, ok := ctx.Value(ContextLogKey).(ctxValue); ok {
		c[key] = value
		return context.WithValue(ctx, ContextLogKey, c)
	}

	return context.WithValue(ctx, ContextLogKey, ctxValue{key: value})
}

func WithLogContextMap(ctx context.Context, data map[string]interface{}) context.Context {
	if c, ok := ctx.Value(ContextLogKey).(ctxValue); ok {
		for k, v := range data {
			c[k] = v
		}
		return context.WithValue(ctx, ContextLogKey, c)
	}

	return context.WithValue(ctx, ContextLogKey, ctxValue(data))
}
