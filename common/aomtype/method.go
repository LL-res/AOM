package aomtype

import (
	"errors"
)

func (h *Hide) Init() {
	h.MetricMap.NewConcurrentMap()
	h.PredictorMap.NewConcurrentMap()
	h.ModelMap.NewConcurrentMap()
	h.CollectorWorkerMap.NewConcurrentMap()
}

func (m *Map[T]) New() {
	m.Data = make(map[string]T)
}

func (m *Map[T]) Get(key string) (T, error) {
	val, ok := m.Data[key]
	if !ok {
		return val, errors.New("value not found")
	}
	return val, nil
}

func (m *Map[T]) Delete(key string) {
	delete(m.Data, key)
}

func (m *Map[T]) Set(key string, value T) {
	m.Data[key] = value
}
