package event

import (
	"fmt"
	"go-im/pkg/logger"
	"reflect"
	"sync"
)

type eventHandler struct {
	callBack reflect.Value
}

type IBus interface {
	Subscribe(topic string, handler any) error
	Publish(topic string, args ...any)
}

type AsyncEventBus struct {
	handlers map[string][]*eventHandler
	lock     sync.Mutex
}

func NewAsyncEventBus() *AsyncEventBus {
	return &AsyncEventBus{
		handlers: make(map[string][]*eventHandler),
		lock:     sync.Mutex{},
	}
}

// Subscribe 订阅
func (b *AsyncEventBus) Subscribe(topic string, f any) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !(reflect.TypeOf(f).Kind() == reflect.Func) {
		return fmt.Errorf("%s is not of type reflect.Func", reflect.TypeOf(f).Kind())
	}

	b.handlers[topic] = append(b.handlers[topic], &eventHandler{
		callBack: reflect.ValueOf(f),
	})
	return nil
}

// Publish 发布
// 这里异步执行，并且不会等待返回结果
func (b *AsyncEventBus) Publish(topic string, args ...any) {
	handlers, ok := b.handlers[topic]
	if !ok {
		logger.Error("[event bus] 未注册订阅，topic：" + topic)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(handlers))

	// 如果有 Unsubscribe 或 removeHandler 等操作，需要拷贝处理函数，防止执行过程中被移除
	//copyHandlers := make([]*eventHandler, len(handlers))
	//copy(copyHandlers, handlers)

	for i, handler := range handlers {
		go func(i int) {
			defer wg.Done()
			passedArguments := b.setUpPublish(handler, args...)
			handler.callBack.Call(passedArguments)
		}(i)
	}
	wg.Wait()
}

func (b *AsyncEventBus) setUpPublish(callback *eventHandler, args ...interface{}) []reflect.Value {
	funcType := callback.callBack.Type()
	passedArguments := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			passedArguments[i] = reflect.New(funcType.In(i)).Elem()
		} else {
			passedArguments[i] = reflect.ValueOf(v)
		}
	}

	return passedArguments
}
