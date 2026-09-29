package http

import (
	"context"
	"uuid"

	"github.com/quangbach27/golang-common/log"
)

type ctxKey int

const correlationIDKey ctxKey = iota

func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

func CorrelationIDFromContext(ctx context.Context) string {
	v, ok := ctx.Value(correlationIDKey).(string)
	if ok {
		return v
	}

	log.FromContext(ctx).Warn("correlation ID not found in context")

	// add "gen_" prefix to distinguish generated correlation IDs from correlation IDs passed by the client
	// it's useful to detect if correlation ID was not passed properly
	return "gen_" + uuid.NewV4().String()
}
