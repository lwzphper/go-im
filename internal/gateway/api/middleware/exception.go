package middleware

import (
	"errors"
	"fmt"
	"go-im/pkg/response"
	"go-im/pkg/util"

	"github.com/gin-gonic/gin"
)

func Exception() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				errStr := errorToString(err)
				util.LogError(c, errors.New(errStr))
				response.InternalError(c.Writer)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func errorToString(err interface{}) string {
	switch v := err.(type) {
	// case WrapError: // 自定义异常
	// 符合预期的错误，可以直接返回给客户端
	//	return v.Data
	case error:
		return fmt.Sprintf("panic: %v\n", v.Error())
	default:
		// 同上
		if s, ok := err.(string); ok {
			return s
		}

		return fmt.Sprintf("Unknown error：%#v\n", err)
	}
}
