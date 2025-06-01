package cqrs

import (
	"context"
	"github.com/sirupsen/logrus"
)

func ApplyQueryDecorator[Q any, R any](baseHandler QueryHandler[Q, R], logger *logrus.Entry) QueryHandler[Q, R] {
	return queryLoggingDecorator[Q, R]{
		baseHandler: baseHandler,
		logger:      logger,
	}
}

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
