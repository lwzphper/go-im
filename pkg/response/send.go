package response

import (
	"encoding/json"
	"fmt"
	"go-im/pkg/errorx"
	"go-im/pkg/logger"
	"math"
	"net/http"
)

// Success 成功响应
func Success(w http.ResponseWriter, data any, options ...RespOption) {
	Send(w, data, CodeSuccess, options...)
}

// Dynamic 动态响应数据（根据是否有错误，来决定响应数据）
func Dynamic(w http.ResponseWriter, data any, err error, options ...RespOption) {
	if err != nil {
		errorResponse(w, data, err, options...)
	} else {
		Success(w, data)
	}
}

// DynamicPage 动态响应数据（根据是否有错误，来决定响应数据）
func DynamicPage(w http.ResponseWriter, data any, page, pageSize, total int, err error, options ...RespOption) {
	if err != nil {
		errorResponse(w, data, err, options...)
	} else {
		PageSuccess(w, data, page, pageSize, total, options...)
	}
}

func Error(w http.ResponseWriter, err error, options ...RespOption) {
	errorResponse(w, nil, err, options...)
}

// 返回错误响应
func errorResponse(w http.ResponseWriter, data any, err error, options ...RespOption) {
	errx := errorx.FromError(err)
	options = append(options, WithMsg(errx.Message))
	Send(w, data, errx.Code, options...)
}

// PageSuccess 分页响应数据
func PageSuccess(w http.ResponseWriter, data any, page, pageSize, total int, options ...RespOption) {
	options = append(options, func(r *Response) {
		pageOptions := PageOptions{
			CurrentPage: page,
			PerPage:     pageSize,
		}
		if total > 0 {
			pageOptions.Total = total
			pageOptions.TotalPages = int(math.Ceil(float64(total) / float64(pageSize)))
		}
		if r.Meta == nil {
			r.Meta = make(map[string]any)
		}
		r.Meta["pagination"] = pageOptions
	})
	Send(w, data, CodeSuccess, options...)
}

// UnauthorizedError 未授权
func UnauthorizedError(w http.ResponseWriter, options ...RespOption) {
	options = append(options, WithHttpStatusCode(http.StatusUnauthorized))
	Send(w, nil, CodeUnauthorized, options...)
}

// NotFoundError 页面未找到
func NotFoundError(w http.ResponseWriter) {
	Send(w, nil, CodeNotFound, WithMsg("请求地址有误"), WithHttpStatusCode(http.StatusNotFound))
}

// FormValidError 表单验证错误
func FormValidError(w http.ResponseWriter, msg string) {
	Send(w, nil, CodeInvalidParams, WithMsg(msg), WithHttpStatusCode(http.StatusBadRequest))
}

// InternalError 内部错误
func InternalError(w http.ResponseWriter, options ...RespOption) {
	options = append(options, WithHttpStatusCode(http.StatusInternalServerError))
	Send(w, nil, CodeRuntimeError, options...)
}

// Failed 错误响应
func Failed(w http.ResponseWriter, options ...RespOption) {
	Send(w, nil, CodeDefaultError, options...)
}

// Send 发送响应
func Send(w http.ResponseWriter, data any, code int, options ...RespOption) {
	options = append(options, WithData(data))
	resp := NewResponse(code, options...)
	handleSend(w, resp)
}

// handleSend 发送请求
func handleSend(w http.ResponseWriter, resp *Response) {
	respByt, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := fmt.Sprintf(`{"code":"%d", "msg": "encoding to json error, %s"}`, CodeRuntimeError, err)

		if _, err = w.Write([]byte(errMsg)); err != nil {
			logger.Error("send response error: " + err.Error()) // 错误默认输出到终端

			return
		}

		return
	}

	// set response header
	w.Header().Set("Content-Type", "application/json")

	for key, val := range resp.headers {
		w.Header().Set(key, val)
	}

	w.WriteHeader(resp.httpStatusCode)

	if _, err = w.Write(respByt); err != nil {
		logger.Error("send response error: " + err.Error()) // 错误默认输出到终端

		return
	}
}
