package response

import (
	"net/http"
)

func NewResponse(code int, options ...RespOption) *Response {
	resp := &Response{
		Data:           nil,
		Status:         code,
		httpStatusCode: http.StatusOK,
		headers:        map[string]string{},
	}

	for _, option := range options {
		option(resp)
	}

	// http 状态码先统一返回 200
	resp.httpStatusCode = http.StatusOK

	if len(resp.Msg) == 0 {
		resp.Msg = GetMsg(resp.Status)
	}

	return resp
}

type Response struct {
	Status         int            `json:"status"`
	Msg            string         `json:"message"`
	Data           any            `json:"data"`
	Meta           map[string]any `json:"meta,omitempty"`
	httpStatusCode int
	headers        map[string]string
}

// CheckError 检查是否错误
func (r *Response) CheckError() bool {
	return r.Status == CodeSuccess
}

type PageOptions struct {
	Total       int `json:"total,omitempty"`
	PerPage     int `json:"per_page,omitempty"`
	CurrentPage int `json:"current_page,omitempty"`
	TotalPages  int `json:"total_pages,omitempty"`
}

type RespOption func(*Response)

// WithMetaData 设置元数据
func WithMetaData(key string, value any) RespOption {
	return func(r *Response) {
		if r.Meta == nil {
			r.Meta = make(map[string]any)
		}

		r.Meta[key] = value
	}
}

// WithData 设置数据
func WithData(data any) RespOption {
	return func(r *Response) {
		r.Data = data
	}
}

// WithMsg 设置提示信息
func WithMsg(msg string) RespOption {
	return func(r *Response) {
		r.Msg = msg
	}
}

func WithStatus(status int) RespOption {
	return func(r *Response) {
		r.Status = status
	}
}

func WithHeaders(headers map[string]string) RespOption {
	return func(r *Response) {
		r.headers = headers
	}
}

func WithAuthHeader(token string) RespOption {
	return func(r *Response) {
		r.headers["Authorization"] = token
	}
}

func WithHttpStatusCode(code int) RespOption {
	return func(r *Response) {
		r.httpStatusCode = code
	}
}
