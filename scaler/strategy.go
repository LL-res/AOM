package scaler

import (
	"github.com/LL-res/AOM/utils"
	"math"
)

// 决定如何将指标参数转化为副本数
type BaseStrategy func(targetMetric, startMetric float64, startReplica int32, predictMetric []float64) []int32

// 决定如何将多个model的预测副本数转化统一成一个作为metric的副本数
type ModelStrategy func(replicas [][]int32) []int32

// 决定如何将多个指标的副本数统一成一个预测副本数，并将其作为最终监控对象的预测副本数
type MetricStrategy func(replicas [][]int32) []int32

type ObjStrategy func(replicas []int32) int32

func Steady(targetMetric, startMetric float64, startReplica int32, predictMetric []float64) []int32 {
	res := make([]int32, 0)
	for _, m := range predictMetric {
		if m < 0 {
			res = append(res, 0)
			continue
		}
		//向上取整
		res = append(res, int32(math.Ceil(float64(startReplica)*(m/startMetric))))
	}
	return res
}

func UnderThreshold(targetMetric, startMetric float64, startReplica int32, predictMetric []float64) []int32 {
	res := make([]int32, 0)
	for _, m := range predictMetric {
		if m < 0 {
			res = append(res, 0)
			continue
		}
		res = append(res, int32(math.Ceil(m/targetMetric)))
		//if m >= targetMetric {
		//	if i == 0 {
		//		res = append(res, int32(math.Floor(float64(startReplica)*(m/targetMetric))))
		//		continue
		//	}
		//	res = append(res, int32(math.Floor(float64(res[i-1])*(m/targetMetric))))
		//	continue
		//}
		//if i == 0 {
		//	res = append(res, startReplica)
		//	continue
		//}
		//res = append(res, res[i-1])
	}
	return res
}

func MaxStrategy(replicas [][]int32) []int32 {
	if len(replicas) == 0 {
		return nil
	}
	res := make([]int32, len(replicas[0]))
	for _, v := range replicas {
		for i, vv := range v {
			res[i] = utils.Max(vv, res[i])
		}
	}
	return res
}

func MinStrategy(replicas [][]int32) []int32 {
	if len(replicas) == 0 {
		return nil
	}
	res := make([]int32, len(replicas[0]))
	copy(res, replicas[0])
	for idx, v := range replicas {
		if idx == 0 {
			continue
		}
		for i, vv := range v {
			res[i] = utils.Min(vv, res[i])
		}
	}
	return res
}

func MeanStrategy(replicas [][]int32) []int32 {
	if len(replicas) == 0 {
		return nil
	}
	res := make([]int32, len(replicas[0]))
	for _, v := range replicas {
		for i, vv := range v {
			res[i] += vv
		}
	}
	for i := range res {
		res[i] /= int32(len(replicas))
	}
	return res
}

func SelectMax(replicas []int32) int32 {
	return utils.Max(replicas...)
}
