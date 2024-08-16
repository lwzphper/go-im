package response

const (
	// 全局

	CodeSuccess       int = 0
	CodeRuntimeError  int = 500
	CodeInvalidParams int = 400
	CodeUnauthorized  int = 30000 // 兼容旧版接口
	CodeNotFound      int = 404
	CodeDefaultError  int = 800

	// 其他模块自定义
)

var CodeMsgMap = map[int]string{
	CodeSuccess:       "成功",
	CodeRuntimeError:  "服务器繁忙，请稍后再试。",
	CodeInvalidParams: "表单验证有误",
	CodeUnauthorized:  "授权失败，请求重新登录。",
	CodeDefaultError:  "添加数据失败",
	CodeNotFound:      "数据不存在",
}

func GetMsg(c int) string {
	if msg, ok := CodeMsgMap[c]; ok {
		return msg
	}
	return "未知"
}
