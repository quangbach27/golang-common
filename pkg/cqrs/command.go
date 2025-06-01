package cqrs

import (
	"context"
	"github.com/sirupsen/logrus"
)

func ApplyCommandDecorators[C any](baseHandler CommandHandler[C], logger *logrus.Entry) CommandHandler[C] {
	return commandLoggingDecorator[C]{
		baseHandler: baseHandler,
		logger:      logger,
	}
}

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}
