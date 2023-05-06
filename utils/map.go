package utils

import "fmt"

type Map[T any] struct {
	Data map[string]T
}

func (m *Map[T]) New() {
	m.Data = make(map[string]T)
}

func (m *Map[T]) Get(key string) (T, error) {
	val, ok := m.Data[key]
	if !ok {
		return val, fmt.Errorf("value not found,key [%s]", key)
	}
	return val, nil
}

func (m *Map[T]) Delete(key string) {
	delete(m.Data, key)
}

func (m *Map[T]) Set(key string, value T) {
	m.Data[key] = value
}
