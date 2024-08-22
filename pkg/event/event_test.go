package event

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func msg1(name string, age uint8) {
	fmt.Printf("name: %s, age: %d\n", name, age)
}

func msg2() {
	fmt.Println("hello world")
}

func TestEventBus(t *testing.T) {
	var bus = NewAsyncEventBus()
	cases := []struct {
		name     string
		Topic    string
		handleFn any
		args     []any
	}{
		{
			name:     "msg1 normal",
			Topic:    "msg1",
			handleFn: msg1,
			args:     []any{"john", uint8(18)},
		},
		{
			name:     "msg2 normal",
			Topic:    "msg2",
			handleFn: msg2,
			args:     []any{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := bus.Subscribe(c.Topic, c.handleFn)
			assert.Nil(t, err)
			bus.Publish(c.Topic, c.args...)
		})
	}
}
