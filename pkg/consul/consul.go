package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

type Client struct {
	client  *api.Client
	Address string
}

// ServiceList 服务信息列表
type ServiceList struct {
	Address string
	Port    int
}

// 格式化代理列表
func (s *ServiceList) format(ag *api.AgentService) *ServiceList {
	if ag == nil {
		return nil
	}
	return &ServiceList{
		Address: ag.Address,
		Port:    ag.Port,
	}
}

func NewClient(address string) *Client {
	c := api.DefaultConfig()
	c.Address = address
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}

	return &Client{client: client, Address: address}
}

// Register 服务注册
func (c *Client) Register(host string, port int, regName, serverId string, tags []string, opts ...CheckOption) error {
	registration := new(api.AgentServiceRegistration)
	registration.ID = serverId  // 服务节点的名称
	registration.Name = regName // 服务名称
	registration.Port = port    // 服务端口
	registration.Tags = tags    // tag，可以为空
	registration.Address = host // 服务 IP 要确保consul可以访问这个ip

	// 增加consul健康检查回调函数
	check := &api.AgentServiceCheck{
		Interval:                       "5s",
		Timeout:                        "5s",
		HTTP:                           fmt.Sprintf("http://%s:%d/health", registration.Address, registration.Port),
		DeregisterCriticalServiceAfter: "1m",
	}
	for _, opt := range opts {
		opt(check)
	}
	registration.Check = check

	// 注册服务到consul
	return c.client.Agent().ServiceRegister(registration)
}

// DeRegister 取消注册
func (c *Client) DeRegister(serverId string) error {
	return c.client.Agent().ServiceDeregister(serverId)
}

// AllService 获取全部服务
func (c *Client) AllService() (map[string]*api.AgentService, error) {
	// 获取所有service
	return c.client.Agent().Services()
}

// HealthService 获取健康的节点
func (c *Client) HealthService(name, tag string) (result []*api.AgentService, err error) {
	entry, _, err := c.client.Health().Service(name, tag, false, nil)
	if err != nil {
		return result, err
	}

	for _, e := range entry {
		if e.Checks.AggregatedStatus() == api.HealthPassing {
			result = append(result, e.Service)
		}
	}

	return result, nil
}

// ServiceByName 通过服务名称获取服务列表
func (c *Client) ServiceByName(name string) ([]*api.AgentService, error) {
	var result []*api.AgentService
	services, err := c.client.Agent().ServicesWithFilter(fmt.Sprintf("Service==`%s`", name))
	if err != nil {
		return result, err
	}

	for _, service := range services {
		result = append(result, service)
	}

	return result, nil
}

// Service 获取服务列表
func (c *Client) Service(serverId string) (*api.AgentService, error) {
	// 获取指定service
	service, _, err := c.client.Agent().Service(serverId, nil)
	return service, err
}

// CheckHealth 健康检查
func (c *Client) CheckHealth(serverId string) {
	a, b, _ := c.client.Agent().AgentHealthServiceByID(serverId)
	fmt.Println("val1:", a)
	fmt.Println("val2:", b)
	fmt.Println("ConsulCheckHeath done")
}

// SetKey 设置键值
func (c *Client) SetKey(key, val string) error {
	_, err := c.client.KV().Put(&api.KVPair{Key: key, Value: []byte(val)}, nil)
	return err
}

// GetKey 获取键值
func (c *Client) GetKey(key string) (string, error) {
	pair, _, err := c.client.KV().Get(key, nil)
	result := ""
	if pair != nil {
		result = string(pair.Value)
	}
	return result, err
}

// ListKey 获取列表
func (c *Client) ListKey(key string) ([]string, error) {
	keys, _, err := c.client.KV().Keys(key, "", nil)
	return keys, err
}
