package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-im/pkg/logger"
	"io"
	"net/http"
	"runtime/debug"
)

// LogError 记录错误信息
// data 参数为请求参数
func LogError(c context.Context, err error) {
	logger.Error(getErrorLogInfo(c, err))
}

// getErrorLogInfo 获取错误信息
func getErrorLogInfo(c context.Context, err error) string {
	if c != nil {
		if ginCtx, ok := c.(*gin.Context); ok {
			return getGinCtxInfo(ginCtx, err)
		}

	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	// 普通方式
	return fmt.Sprintf("%s\n错误等级：%s\nstatck:%s", errMsg, "error", debug.Stack())
}

// 获取 gin 上下文信息
func getGinCtxInfo(c *gin.Context, err error) string {
	var (
		postData   []byte
		marshalErr error
	)

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	if c.Request.Method != http.MethodGet {
		if c.ContentType() == gin.MIMEJSON {
			defer c.Request.Body.Close()
			postData, _ = io.ReadAll(c.Request.Body)
		} else {
			params, parseErr := getPostFormParams(c)
			if parseErr == nil {
				if postData, marshalErr = json.Marshal(params); marshalErr != nil {
					errMsg += "post json Marshal error：" + marshalErr.Error()
				}
			}
		}
	}

	return fmt.Sprintf(
		"错误信息：%s\n访问地址：%s\n访问方式：%s\n错误等级：%s\npost请求数据：%s\nstatck:%s",
		errMsg, c.Request.Host+c.Request.URL.String(), c.Request.Method, "error", postData, debug.Stack(),
	)
}

// 获取post请求数据
func getPostFormParams(c *gin.Context) (map[string]any, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return nil, err
		}
	}
	var postMap = make(map[string]any, len(c.Request.PostForm))

	for k, v := range c.Request.PostForm {
		if len(v) > 1 {
			postMap[k] = v
		} else if len(v) == 1 {
			postMap[k] = v[0]
		}
	}

	return postMap, nil
}
