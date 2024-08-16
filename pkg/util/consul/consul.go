package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"go-im/pkg/consul"
	"go-im/pkg/logger"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

const (
	consulRegName = "go-im-service"
	consulTagName = "go-im"
)

var consulClient *Consul

type WatchCallback func([]*api.AgentService) // 服务变更回调，参数为：存活的服务

func C() *Consul {
	if consulClient == nil {
		panic("请先初始化 consul")
	}
	return consulClient
}

// Init 初始化 consul
func Init(addr string) {
	consulClient = &Consul{
		client:          consul.NewClient(addr),
		healthList:      make([]*api.AgentService, 0),
		serviceErrNum:   make(map[string]int),
		allowErrNum:     2,
		checkHealthTime: 1 * time.Minute,
		lastIndex:       0,
	}
}

type Consul struct {
	client          *consul.Client
	healthList      []*api.AgentService
	watchNotify     chan struct{}
	serviceErrNum   map[string]int // 服务心跳失败次数
	allowErrNum     int            // 运行失败的次数
	checkHealthTime time.Duration  // 心跳检查间隔时间

	lastIndex int // 当前节点的索引
}

// ServerRegister 服务注册
// @param host 服务地址
// @param port 服务端口
// @param serverId 服务ID
func (c *Consul) ServerRegister(host string, port int, serverId string) error {
	err := c.client.Register(host, port, consulRegName, serverId, []string{consulTagName})
	if err != nil {
		logger.Error("register to consul error", zap.Error(err), zap.String("address", fmt.Sprintf("%s:%d", host, port)))

		return err
	}

	return nil
}

// HealthService 获取健康的节点信息
func (c *Consul) HealthService() ([]*api.AgentService, error) {
	result, err := c.client.HealthService(consulRegName, consulTagName)
	if err != nil {
		return result, err
	}

	for _, service := range result {
		service.Address = c.getServerAddress(service.Address)
	}
	return result, nil
}

// DeRegister 服务注销
func (c *Consul) DeRegister(serverId string) {
	logger.Debug("注销服务：" + serverId)
	c.removeService(serverId)
	err := c.client.DeRegister(serverId)
	if err != nil {
		logger.Error("deregister to consul error", zap.Error(err), zap.String("server_id", serverId))
	}
}

// 移除节点信息
func (c *Consul) removeService(serverId string) {
	c.removeHealthService(serverId)
	delete(c.serviceErrNum, serverId)
}

// 移除节点
func (c *Consul) removeHealthService(serverId string) {
	for i := 0; i < len(c.healthList); i++ {
		if c.healthList[i].ID == serverId {
			c.healthList = append(c.healthList[:i], c.healthList[i+1:]...)
			i--
			logger.Debug("移除节点：" + serverId)
		}
	}
}

// RoundHealthServerUrl RoundHealthServer 轮询获取健康的服务节点
func (c *Consul) RoundHealthServerUrl() string {
	var err error
	// 没有节点，主动请求 consul 获取
	if len(c.healthList) == 0 {
		if c.healthList, err = c.HealthService(); err != nil {
			logger.Error("get consul service error", zap.Error(err))
			return ""
		}
	}

	// 主动也获取不到节点，就直接返回
	if len(c.healthList) == 0 {
		return ""
	}

	// 防止剔除无效节点，索引越界
	if c.lastIndex+1 > len(c.healthList) {
		c.lastIndex = 0
	}

	service := c.healthList[c.lastIndex]

	c.lastIndex++

	return fmt.Sprintf("%s:%d", c.getServerAddress(service.Address), service.Port)
}

// 兼容 docker 地址
func (c *Consul) getServerAddress(addr string) string {
	if addr == "host.docker.internal" {
		return "127.0.0.1"
	}
	return addr
}

// Heartbeat 心跳检查健康节点（测试中发现，服务注销时，有时候没有通知，这里是兜底方案）
func (c *Consul) Heartbeat() {
	ticker := time.NewTicker(c.checkHealthTime)
	for range ticker.C {
		for _, service := range c.healthList {
			host := fmt.Sprintf("http://%s:%d", c.getServerAddress(service.Address), service.Port)
			healthSuccess := c.checkHealth(host)
			if healthSuccess {
				//logger.Debug("服务心跳成功：" + service.ID + "，地址：" + host)
				c.serviceErrNum[service.ID] = 0
				continue
			}

			logger.Debug("服务心跳失败：" + service.ID + "，地址：" + host)
			// 处理心跳检查失败的情况
			errNum, ok := c.serviceErrNum[service.ID]
			if !ok {
				errNum = 0
			}

			if errNum > c.allowErrNum {
				c.removeService(service.ID)
				continue
			}

			c.serviceErrNum[service.ID] = errNum + 1
		}
	}
}

// 检查节点是否健康
func (c *Consul) checkHealth(host string) bool {
	var transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 连接超时时间
			KeepAlive: 60 * time.Second, // 保持长连接的时间
		}).DialContext, // 设置连接的参数
		MaxIdleConns:          10,               // 最大空闲连接
		IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
		ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
		MaxIdleConnsPerHost:   2,                // 每个host保持的空闲连接数
	}
	client := http.Client{Transport: transport} // 初始化http的client
	resp, err := client.Get(host + "/health")
	if err != nil {
		logger.Error("check health error", zap.Error(err))
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

// Watch 监听服务变化
func (c *Consul) Watch(cb WatchCallback) {
	err := consul.RegisterWatcher(consul.WatchTypeServices, nil, c.client.Address, func(idx uint64, data interface{}) {
		logger.Debug("监听到 consul 节点状态变更")
		var healthList = make([]*api.AgentService, 0)
		switch d := data.(type) {
		case []*api.ServiceEntry:
			for _, i := range d {
				if i.Service.Service == consulRegName && i.Checks.AggregatedStatus() == api.HealthPassing {
					healthList = append(healthList, i.Service)
				}
			}
			c.healthList = healthList
			cb(healthList)
		}
	})
	if err != nil {
		logger.Error("watch consul error", zap.Error(err))
	}
}
