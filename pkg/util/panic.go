package util

import (
	"fmt"
	"go-im/pkg/logger"
	"runtime"

	"go.uber.org/zap"
)

// RecoverPanic 恢复panic
func RecoverPanic() {
	err := recover()
	if err != nil {
		logger.DPanic("panic", zap.Any("panic", err), zap.String("stack", getStackInfo()))
	}
}

// 获取Panic堆栈信息
func getStackInfo() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return fmt.Sprintf("%s", buf[:n])
}
