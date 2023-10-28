package main

import (
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

type LogAdaptor struct {
	logger *zap.Logger
}

func (l *LogAdaptor) Debug(msg string, keyvals ...interface{}) {
	args := append([]any{msg}, keyvals...)
	l.logger.Sugar().Debug(args...)
}

func (l *LogAdaptor) Info(msg string, keyvals ...interface{}) {
	args := append([]any{msg}, keyvals...)
	l.logger.Sugar().Info(args...)
}

func (l *LogAdaptor) Warn(msg string, keyvals ...interface{}) {
	args := append([]any{msg}, keyvals...)
	l.logger.Sugar().Warn(args...)
}

func (l *LogAdaptor) Error(msg string, keyvals ...interface{}) {
	args := append([]any{msg}, keyvals...)
	l.logger.Sugar().Error(args...)
}

var _ log.Logger = (*LogAdaptor)(nil)

func NewLogAdaptor(logger *zap.Logger) *LogAdaptor {
	return &LogAdaptor{logger}
}
