package cache

import (
	"container/list"
	"fmt"
	"sync"
)

type LruCache struct {
	size     int
	values   *list.List
	cacheMap map[any]*list.Element
	lock     sync.Mutex
}

func NewLruList(size int) *LruCache {
	values := list.New()

	return &LruCache{
		size:     size,
		values:   values,
		cacheMap: make(map[any]*list.Element, size),
	}
}

func (l *LruCache) Put(k, v any) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.values.Len() == l.size {
		back := l.values.Back()
		l.values.Remove(back)
		delete(l.cacheMap, back)
	}

	front := l.values.PushFront(v)
	l.cacheMap[k] = front
}

func (l *LruCache) Get(k any) (any, bool) {
	v, ok := l.cacheMap[k]
	if ok {
		l.values.MoveToFront(v)
		return v.Value, true
	} else {
		return nil, false
	}
}

func (l *LruCache) Size() int {
	return l.values.Len()
}
func (l *LruCache) String() {
	for i := l.values.Front(); i != nil; i = i.Next() {
		fmt.Print(i.Value, "\t")
	}
}
func (l *LruCache) List() []any {
	var data []any
	for i := l.values.Front(); i != nil; i = i.Next() {
		data = append(data, i.Value)
	}
	return data
}

func (l *LruCache) Clear() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.values = list.New()
	l.cacheMap = make(map[any]*list.Element, l.size)

}
