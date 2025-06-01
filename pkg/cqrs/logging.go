package cqrs

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
)

type commandLoggingDecorator[C any] struct {
	baseHandler CommandHandler[C]
	logger      *logrus.Entry
}

func (cmdLogDecorator commandLoggingDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)

	logger := cmdLogDecorator.logger.WithFields(logrus.Fields{
		"command_type": handlerType,
		"command_body": fmt.Sprintf("%#v", cmd),
	})

	logger.Debug("Executing command")
	defer func() {
		if err != nil {
			logger.WithError(err).Error("Failed to execute command")
		} else {
			logger.Info("Command executed successfully")
		}
	}()
	return cmdLogDecorator.baseHandler.Handle(ctx, cmd)
}

type queryLoggingDecorator[Q any, R any] struct {
	baseHandler QueryHandler[Q, R]
	logger      *logrus.Entry
}

func (queryLogDecorator queryLoggingDecorator[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	logger := queryLogDecorator.logger.WithFields(logrus.Fields{
		"query_type": generateActionName(query),
		"query_body": fmt.Sprintf("%#v", query),
	})

	logger.Debug("Executing query")
	defer func() {
		if err != nil {
			logger.WithError(err).Error("Failed to execute query")
		} else {
			logger.Info("Query executed successfully")
		}
	}()

	return queryLogDecorator.baseHandler.Handle(ctx, query)
}
