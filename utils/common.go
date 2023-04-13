package utils

import "sync"

type ConcurrentMap[T any] struct {
	data map[string]T
	sync.RWMutex
}

func (m *ConcurrentMap[T]) Load(key string) (T, bool) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.data[key]
	return val, ok
}
func (m *ConcurrentMap[T]) Store(key string, val T) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = val
}
func (m *ConcurrentMap[T]) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
}
func NewConcurrentMap[T any]() *ConcurrentMap[T] {
	return &ConcurrentMap[T]{
		data: make(map[string]T),
	}
}
