package connect

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-im/config"
	"go-im/pkg/logger"
	"go.uber.org/zap"
	"time"
)

var gatewayClient *websocket.Conn

type GatewayClient struct {
	conn *websocket.Conn
}

// 获取客户端
func getGatewayClient() *websocket.Conn {
	var err error
	if gatewayClient == nil {
		gatewayClient = gatewayDail()
		return gatewayClient
	}

	// 判断连接是否可用
	if err = gatewayClient.WriteControl(websocket.PingMessage, nil, time.Now().Add(1*time.Second)); err != nil {
		gatewayClient.Close()
		if d := gatewayDail(); d != nil { // 重试连接一次
			return d
		}
		return nil
	}

	return gatewayClient
}

// 尝试连接
func gatewayDail() *websocket.Conn {
	var err error
	authKey := fmt.Sprintf("?%s=%s", config.GatewayAuthKey, config.GatewayAuthVal)
	logger.Debug("网关地址", zap.String("addr", config.C.GetGatewayWsAddr()))
	gatewayClient, _, err = websocket.DefaultDialer.Dial(config.C.GetGatewayWsAddr()+authKey, nil)
	if err != nil {
		logger.Error("gateway ws dial error", zap.Error(err))
		return nil
	}
	return gatewayClient
}
