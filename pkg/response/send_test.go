package response

import (
	"fmt"
	"github.com/pkg/errors"
	"go-im/pkg/errorx"
	"net/http"
	"testing"
)

type response struct {
}

func (r response) Header() http.Header {
	return http.Header{}
}

func (r response) Write(bytes []byte) (int, error) {
	fmt.Printf(string(bytes))
	return 0, nil
}

func (r response) WriteHeader(statusCode int) {

}

func TestPage(t *testing.T) {
	rsp := response{}
	PageSuccess(rsp, nil, 1, 10, 201, WithMetaData("test", "1111"))
}

func TestError(t *testing.T) {
	err := errorx.New(40001, "企业未认证", "请先完成企业认证")
	err2 := errors.New("11111")

	fErr1 := errorx.FromError(err)
	fErr2 := errorx.FromError(err2)
	fmt.Printf("code:%d msg:%s\n", fErr1.Code, fErr2.Message)
	fmt.Printf("code:%d msg:%s", fErr2.Code, fErr2.Message)
}
