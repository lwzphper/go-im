package consul

import "github.com/hashicorp/consul/api"

type CheckOption func(check *api.AgentServiceCheck)

// WithTimeout 超时时间
func WithTimeout(t string) CheckOption {
	return func(check *api.AgentServiceCheck) {
		check.Timeout = t
	}
}

// WithInterval 健康检查间隔
func WithInterval(t string) CheckOption {
	return func(check *api.AgentServiceCheck) {
		check.Interval = t
	}
}

// WithDeregisterCriticalServiceAfter 故障检查失败，删除服务时间
func WithDeregisterCriticalServiceAfter(t string) CheckOption {
	return func(check *api.AgentServiceCheck) {
		check.DeregisterCriticalServiceAfter = t
	}
}
