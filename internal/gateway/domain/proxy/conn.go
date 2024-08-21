package proxy

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-im/config"
	"go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Proxy struct {
	servers map[string]*websocket.Conn // 已连接的服务
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 处理 ws 请求
func Handle(c *gin.Context) {
	wsConn, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer wsConn.Close()

	// 校验
	if c.Query(config.GatewayAuthKey) != config.GatewayAuthVal {
		logger.Debug("auth key 不合法", zap.String("auth key", c.Query(config.GatewayAuthKey)))
		wsConn.WriteMessage(websocket.TextMessage, types.MarshalOutput(types.MethodServiceNotice, "无权操作", 0))
		return
	}

	logger.Debugf("服务已连接 gateway")

	for {
		var message []byte
		_, message, err = wsConn.ReadMessage()
		/*if err != nil && connect.WsErrorNeedClose(err) {
			_ = wsConn.Close()
			return
		}*/
		if err != nil {
			logger.Debug("proxy 读取消息失败", zap.Error(err))
			return
		}

		var data = new(types.QueueMsgData)
		err = json.Unmarshal(message, data)
		if err != nil {
			logger.Infof("IMServer消息格式有误：%s", string(message))
			_ = wsConn.WriteMessage(websocket.TextMessage, types.MarshalOutput(types.MethodServiceNotice, "消息格式有误", 0))
			continue
		}

		C().Send(data)
	}
}
