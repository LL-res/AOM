package aomtype

import (
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/utils"
)

type Hide struct {
	//noModelKey
	CollectorMap map[string]chan struct{}
	//noModelKey
	MetricMap *utils.ConcurrentMap[*basetype.Metric]
	//withModelKey
	PredictorMap *utils.ConcurrentMap[predictor.Predictor]
	//noModelKey
	CollectorWorkerMap *utils.ConcurrentMap[collector.MetricCollector]
	//withModelKey
	ModelMap *utils.ConcurrentMap[*basetype.Model]
}
