package connect

import (
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"go-im/config"
	"go-im/internal/event"
	"go-im/internal/logic/room/types"
	"go-im/pkg/errorx"
	"go-im/pkg/jwt"
	"go-im/pkg/logger"
	"go-im/pkg/util"
	"go-im/pkg/util/consul"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// 服务连接处理

var (
	ErrAuthenticate = errorx.New(40001, "用户未登录或token无效", "授权失败，请重新登录。")
	ErrHasLogin     = errorx.New(40002, "用户已登录", "用户已登录，请勿重复登录")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 65536,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

type WsConn struct {
	address  string
	serverId string // 服务id，用于 consul 注册
}

func InitServer(addr string) *WsConn {
	c := NewWsConn(addr)
	go c.StartServer()
	return c
}

func NewWsConn(address string) *WsConn {
	guid := xid.New()

	return &WsConn{
		address:  address,
		serverId: config.C.App.Name + "_" + guid.String(),
	}
}

// StartServer 启动服务
func (c *WsConn) StartServer() {
	http.HandleFunc("/ws", c.handleConn)
	http.HandleFunc("/gateway", c.handleGatewayConn)
	// consul 健康检查
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	serverAddr, err := util.SplitAddress(c.address, config.C.App.InDocker)
	if err != nil {
		panic(err)
	}

	// 注册到 consul 中
	if err = consul.C().ServerRegister(serverAddr.Host, serverAddr.Port, c.serverId); err != nil {
		panic(err)
	}

	logger.Infof("ws service start, serverId:%s address: %s", c.serverId, serverAddr.String())
	if err := http.ListenAndServe(c.address, nil); err != nil {
		c.Close()
		panic(err)
	}
}

// Close 服务关闭
func (c *WsConn) Close() {
	// 注销 consul
	consul.C().DeRegister(c.serverId)
}

// 处理连接
func (c *WsConn) handleConn(w http.ResponseWriter, r *http.Request) {
	defer util.RecoverPanic()

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ZapL().Sugar().Error(err)
		return
	}

	// 用户授权
	userId, err := c.auth(r)
	if err != nil {
		OutputError(wsConn, types.CodeAuthError, errorx.Message(err))
		_ = wsConn.Close()
		return
	}

	// 判断是否在当前节点已登录，剔除下线
	if mapNode := GetNode(userId); mapNode != nil {
		OutputError(mapNode.Conn, types.CodeAuthError, "当前账号已被其他用户登录")
		CloseConn(mapNode)
	}

	addr, err := util.SplitAddress(c.address, config.C.App.InDocker)
	if err != nil {
		OutputError(wsConn, types.CodeAuthError, "地址解析失败")
		wsConn.Close()
		return
	}
	node := NewNode(wsConn, userId, addr.String(), c.serverId, WithNodeLoginTime(time.Now().Unix()))

	// 用户跟节点的映射
	SetNode(userId, node)
}

// 处理网关连接
func (c *WsConn) handleGatewayConn(w http.ResponseWriter, r *http.Request) {
	defer util.RecoverPanic()

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ZapL().Sugar().Error(err)
		return
	}

	// 权限判断
	if r.URL.Query().Get(config.GatewayAuthKey) != config.GatewayAuthVal {
		WriteTextMessage(wsConn, types.MethodServiceNotice, "无权操作")
		_ = wsConn.Close()
		return
	}

	logger.Debugf("网关已连接")

	// 消息处理
	for {
		var message []byte
		_, message, err = wsConn.ReadMessage()
		if err != nil {
			_ = wsConn.Close()
			logger.Debug("gateway 读取消息失败", zap.Error(err))
			return
		}

		// 直接绕过空消息，有可能是ping
		if string(message) == "" {
			continue
		}

		logger.Debugf("接收到网关数据：%s", string(message))

		event.RoomEvent.Publish(event.GatewayMsg, wsConn, message)
	}
}

// 授权认证，返回用户id
func (c *WsConn) auth(r *http.Request) (uint64, error) {
	userIdParam := r.URL.Query().Get("user_id")
	if userIdParam != "" { // 调试临时使用
		return util.StringToUint64(userIdParam)
	}

	var token string
	if token = r.Header.Get("Sec-Websocket-Protocol"); token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		logger.Debug("没有传递 token，授权失败")
		return 0, ErrAuthenticate
	}

	validator := jwt.NewTokenValidator([]byte(config.C.Jwt.Secret))
	claims, err := validator.Validator(token)
	if err != nil || claims == nil {
		logger.Debug("token 无效，授权失败", zap.String("token", token), zap.Error(err))
		return 0, ErrAuthenticate
	}

	return claims.Audience, nil
}
