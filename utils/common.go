package utils

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"strings"
)

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
func GetWithModelKey(NoModelKey, model string) string {
	return fmt.Sprintf("%s$%s", NoModelKey, model)
}
func GetNoModelKey(withModelKey string) string {
	strs := strings.Split(withModelKey, "$")
	return strings.Join(strs[:len(strs)-1], "$")
}
func GetModelType(withModelType string) string {
	strs := strings.Split(withModelType, "$")
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
