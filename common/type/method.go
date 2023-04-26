package AOMtype

import (
	"errors"
	"fmt"
)

func (h *Hide) Init() {
	h.MetricMap.New()
	h.PredictorMap.New()
	h.ModelMap.New()
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

func (m Metric) NoModelKey() string {
	return fmt.Sprintf("%s/%s/%s", m.Name, m.Unit, m.Query)
}
func (m Metric) WithModelKey(modelType string) string {
	return fmt.Sprintf("%s/%s", m.NoModelKey(), modelType)
}
