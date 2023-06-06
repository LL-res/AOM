package utils

import (
	"fmt"
	"sync"
)

type ConcurrentMap[T any] struct {
	Data map[string]T
	sync.RWMutex
}

func (m *ConcurrentMap[T]) Load(key string) (T, error) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.Data[key]
	if !ok {
		return val, fmt.Errorf("value not found,key [%s]", key)
	}
	return val, nil
}
func (m *ConcurrentMap[T]) Range(do func(key string, val T, attr ...any) error, attr ...any) (errors []error) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.Data {
		if err := do(k, v, attr...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
func (m *ConcurrentMap[T]) Store(key string, val T) {
	m.Lock()
	defer m.Unlock()
	m.Data[key] = val
}
func (m *ConcurrentMap[T]) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Data, key)
}
func (m *ConcurrentMap[T]) NewConcurrentMap() {
	if m.Data == nil {
		m.Data = make(map[string]T)
	}
}
func NewConcurrentMap[T any]() *ConcurrentMap[T] {
	return &ConcurrentMap[T]{
		Data: make(map[string]T),
	}
}
