package logger

import (
	"go.uber.org/zap"
	"testing"
)

func TestLogger_method(t *testing.T) {
	defer Sync()
	Info("info log")
	Warn("warn log")
	Error("error log")
	DPanic("dPanic log")
}

func TestLogger_resetDefault(t *testing.T) {
	logger := NewByFileName("./test.log", zap.InfoLevel)
	ResetDefault(logger)

	defer Sync()

	Info("info log")
	Warn("warn log")
	Error("error log")
}
