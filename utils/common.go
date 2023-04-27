package utils

import (
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"strings"
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
		return val, errors.New(fmt.Sprint("value not found,key [%s]", key))
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
	if m == nil {
		m = NewConcurrentMap[T]()
	}
}
func NewConcurrentMap[T any]() *ConcurrentMap[T] {
	return &ConcurrentMap[T]{
		Data: make(map[string]T),
	}
}

func Max[T constraints.Ordered](x ...T) T {
	if len(x) == 1 {
		return x[0]
	}
	max := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] > max {
			max = x[i]
		}
	}
	return max
}

func Min[T constraints.Ordered](x ...T) T {
	if len(x) == 1 {
		return x[0]
	}
	min := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] < min {
			min = x[i]
		}
	}
	return min
}
func GetNoModelKey(withModelKey string) string {
	strs := strings.Split(withModelKey, "/")
	return strings.Join(strs[:len(strs)-1], "/")
}
func GetModelType(withModelType string) string {
	strs := strings.Split(withModelType, "/")
	return strs[len(strs)-1]
}
func MulSlice[T constraints.Float | constraints.Integer](k T, nums []T) {
	for i := range nums {
		nums[i] *= k
	}
}
func AddSlice[T constraints.Float | constraints.Integer](nums ...[]T) []T {
	res := make([]T, len(nums[0]))
	for i, num := range nums {
		res[i] += num[i]
	}
	return res
}
