package ptype

import "github.com/LL-res/AOM/collector"

type Base struct {
	MetricHistory []collector.Metric // 存储着全部
	//socket client
}
type PredictResult struct {
	StartMetric   float64
	Loss          float64
	StartReplica  int32
	PredictMetric []float64
}
