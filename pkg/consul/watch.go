package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api/watch"
	"sync"
)

// 监听类型：https://developer.hashicorp.com/consul/docs/dynamic-app-config/watches#watch-types
const WatchTypeKey = "key"             // Watch a specific KV pair
const WatchTypeKeyPrefix = "keyprefix" // Watch a prefix in the KV store
const WatchTypeServices = "services"   // Watch the list of available services
const WatchTypeNodes = "nodes"         // Watch the list of nodes
const WatchTypeService = "service"     // Watch the instances of a service
const WatchTypeChecks = "checks"       // Watch the value of health checks
const WatchTypeEvent = "event"         // Watch for custom user events

// 參考：https://juejin.cn/post/6984378158347157512

// watch包的使用方法为：1）使用watch.Parse(查询参数)生成Plan，2）绑定Plan的handler，3）运行Plan

// 定义watcher
type Watcher struct {
	Address  string                 // consul agent 的地址："127.0.0.1:8500"
	Wp       *watch.Plan            // 总的Services变化对应的Plan
	watchers map[string]*watch.Plan // 对已经进行监控的service作个记录
	RWMutex  *sync.RWMutex
}

// 将consul新增的service加入，并监控
func (w *Watcher) registerServiceWatcher(serviceName string, handler watch.HandlerFunc) error {
	// watch endpoint 的请求参数，具体见官方文档：https://www.consul.io/docs/dynamic-app-config/watches#service
	wp, err := watch.Parse(map[string]interface{}{
		"type":    WatchTypeService,
		"service": serviceName,
	})
	if err != nil {
		return err
	}

	// 定义service变化后所执行的程序(函数)handler
	wp.Handler = handler
	// 启动监控
	go wp.Run(w.Address)
	// 对已启动监控的service作一个记录
	w.RWMutex.Lock()
	w.watchers[serviceName] = wp
	w.RWMutex.Unlock()

	return nil
}

func NewWatcher(watchType string, opts map[string]string, consulAddr string, handler watch.HandlerFunc) (*Watcher, error) {
	var options = map[string]interface{}{
		"type": watchType,
	}
	// 组装请求参数。(监控类型不同，其请求参数不同)
	for k, v := range opts {
		options[k] = v
	}

	wp, err := watch.Parse(options)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Address:  consulAddr,
		Wp:       wp,
		watchers: make(map[string]*watch.Plan),
		RWMutex:  new(sync.RWMutex),
	}

	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		// 这里只实现了对services的监控，其他监控的data类型判断参考：https://github.com/dmcsorley/avast/blob/master/consul.go
		// services
		case map[string][]string:
			for i := range d {
				// 如果该service已经加入到ConsulRegistry的services里监控了，就不再加入 或者i 为 "consul"的字符串
				// 为什么会多一个consul，参考官方文档services监听的返回值：https://www.consul.io/docs/dynamic-app-config/watches#services
				if _, ok := w.watchers[i]; ok || i == "consul" {
					continue
				}
				if err = w.registerServiceWatcher(i, handler); err != nil {
					fmt.Println(err)
				}
			}

			// 从总的services变化中找到不再监控的service并停止
			w.RWMutex.RLock()
			watches := w.watchers
			w.RWMutex.RUnlock()

			// remove unknown services from watchers
			for i, svc := range watches {
				if _, ok := d[i]; !ok {
					svc.Stop()
					delete(watches, i)
				}
			}
		default:
			fmt.Printf("不能判断监控的数据类型: %v", &d)
		}
	}

	return w, nil
}

func RegisterWatcher(watchType string, opts map[string]string, consulAddr string, handler watch.HandlerFunc) error {
	w, err := NewWatcher(watchType, opts, consulAddr, handler)
	if err != nil {
		return err
	}
	defer w.Wp.Stop()
	if err = w.Wp.Run(consulAddr); err != nil {
		return err
	}

	return nil
}
