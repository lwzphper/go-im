package proxy

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/consul/api"
	"go-im/config"
	"go-im/internal/logic/room/types"
	"go-im/pkg/logger"
	"go-im/pkg/util/consul"
	"go.uber.org/zap"
	"time"
)

const (
	MaxChannelSize      = 1000
	pingPeriod          = 60 * time.Second // 心跳时间周期
	connAllowErrorTimes = 3                // 连接允许心跳失败的次数
)

var client *MsgProxy

func C() *MsgProxy {
	if client == nil {
		panic("请先初始化 Proxy")
	}
	return client
}

type connInfo struct {
	conn       *websocket.Conn
	service    *api.AgentService
	errorTimes uint8 // 错误次数
}

// 获取IM健康服务连接
func getHealthImServiceConn() map[string]*connInfo {
	// 获取全部健康服务
	services, err := consul.C().HealthService()
	if err != nil {
		panic(err)
	}

	var conns = make(map[string]*connInfo)
	for _, service := range services {
		var conn *websocket.Conn
		if conn, err = connectService(service); err != nil || conn == nil {
			continue
		}
		conns[service.ID] = &connInfo{
			conn:    conn,
			service: service,
		}
	}
	return conns
}

func Init() {
	client = &MsgProxy{
		conns:         getHealthImServiceConn(),
		serviceChange: make(chan []*api.AgentService, 1),
		proxyMsg:      make(chan *types.QueueMsgData, MaxChannelSize),
	}

	go client.Write()
}

type MsgProxy struct {
	conns         map[string]*connInfo
	serviceChange chan []*api.AgentService // 服务器变更通知
	proxyMsg      chan *types.QueueMsgData // 代理消息
}

// Send 发送消息
func (p *MsgProxy) Send(data *types.QueueMsgData) {
	p.proxyMsg <- data
}

// ServiceChange 服务变更通知
func (p *MsgProxy) ServiceChange(data []*api.AgentService) {
	p.serviceChange <- data
}

// 发送数据
func (p *MsgProxy) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case data := <-p.proxyMsg: // 消息代理
			msgStr := data.Marshal()
			logger.Debug("接收收到 proxy 消息", zap.Int("当前连接数量", len(p.conns)), zap.String("msg", string(msgStr)))
			for serverId, c := range p.conns {
				// 服务转发消息，不转发给发送方
				if serverId == data.FromServer {
					continue
				}

				logger.Debug("发送 proxy 消息", zap.String("ServerId", serverId))
				err := c.conn.WriteMessage(websocket.TextMessage, msgStr)
				if err != nil {
					logger.Error("proxy 消息发送失败，serverId：" + serverId)

					_ = c.conn.Close()
					// 消息发送失败，重新连接
					if conn, err := connectService(c.service); err != nil {
						p.conns[c.service.ID] = &connInfo{conn: conn, service: c.service}
						_ = conn.WriteMessage(websocket.TextMessage, msgStr)
					}
				}
			}
		case services := <-p.serviceChange: // consul 心跳，服务变更通知
			fmt.Println("监听到 consul 节点变更")

			// 记录存活的服务id
			var serviceIds = make(map[string]struct{}, len(services))

			// 连接新增的服务
			for _, service := range services {
				serviceIds[service.ID] = struct{}{}
				if _, ok := p.conns[service.ID]; !ok {
					conn, err := connectService(service)
					if err != nil {
						logger.Debug("[proxy] 连接 IM service 失败", zap.String("serviceId", service.ID), zap.Error(err))
						continue
					}
					logger.Debug("新增服务", zap.String("serviceId", service.ID))
					p.conns[service.ID] = &connInfo{conn: conn, service: service}
				}
			}

			// 移除不存在的服务
			for serviceId, _ := range p.conns {
				if _, ok := serviceIds[serviceId]; !ok {
					p.removeConn(serviceId)
				}
			}
			logger.Debug("当前连接数量", zap.Int("当前连接数量", len(p.conns)))
		case <-ticker.C:
			for serverId, c := range p.conns {
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.errorTimes++
					logger.Error("ping", zap.Error(err))
					if c.errorTimes > connAllowErrorTimes {
						p.removeConn(serverId)
						logger.Error("close connect", zap.String("serverId", serverId), zap.Error(err))
					}
				} else {
					c.errorTimes = 0
				}
			}
		}
	}
}

// 移除连接
func (p *MsgProxy) removeConn(serverId string) {
	if c, ok := p.conns[serverId]; ok {
		_ = c.conn.Close()
		delete(p.conns, serverId)
	}
}

// 连接服务
func connectService(service *api.AgentService) (*websocket.Conn, error) {
	authKeyQuery := fmt.Sprintf("?%s=%s", config.GatewayAuthKey, config.GatewayAuthVal)
	addrs := fmt.Sprintf("ws://%s:%d/gateway", service.Address, service.Port)
	conn, _, err := websocket.DefaultDialer.Dial(addrs+authKeyQuery, nil)
	if err != nil {
		//_ = conn.Close()
		logger.Error("proxy connect service ws dial error", zap.Error(err))
		return nil, err
	}
	return conn, nil
}
