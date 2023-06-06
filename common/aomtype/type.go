package aomtype

import (
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/utils"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type Hide struct {
	//noModelKey
	//use to close collector dynamically
	CollectorMap map[string]chan struct{}
	//noModelKey
	MetricMap utils.ConcurrentMap[*basetype.Metric]
	//withModelKey
	PredictorMap utils.ConcurrentMap[predictor.Predictor]
	//noModelKey
	CollectorWorkerMap utils.ConcurrentMap[collector.MetricCollector]
	//withModelKey
	ModelMap utils.ConcurrentMap[*basetype.Model]
	//withModelKey
	//store the latest timestamp the model trained
	TrainHistory utils.ConcurrentMap[time.Time]
}

type AOMStore map[types.NamespacedName]*Hide
