package AOMtype

import (
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/scheduler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Map[T any] struct {
	Data map[string]T
}
type Hide struct {
	//noModelKey
	CollectorMap map[string]chan struct{}
	//noModelKey
	MetricMap Map[*Metric]
	//withModelKey
	PredictorMap Map[predictor.Predictor]
	//noModelKey
	CollectorWorkerMap Map[collector.MetricCollector]
	//withModelKey
	ModelMap  Map[*Model]
	scheduler scheduler.Scheduler
}
type Metric struct {
	ScaleDownConf ScaleDownConf `json:"scaleDownConf"`
	Target        float64       `json:"target"`
	Name          string        `json:"name"`
	Unit          string        `json:"unit"`
	Query         string        `json:"query"`
}
type ScaleDownConf struct {
	Threshold float64       `json:"threshold"`
	Duration  time.Duration `json:"duration"`
}
type Model struct {
	Type            string           // GRU LSTM
	PredcitInterval *metav1.Duration `json:"predcitInterval"`
	GRU             GRU
	LSTM            LSTM
}
type LSTM struct {
}
type GRU struct {
	// how far in second GRU will use to train
	// +optional
	TrainSize   int    `json:"trainSize"`
	LookBack    int    `json:"lookBack"`
	LookForward int    `json:"lookForward"`
	Address     string `json:"address"`

	//暂时把它当作，需要维持在的值
	ScaleUpThreshold float64 `json:"scaleUpThreshold"`
	//retrain interval
	UpdateInterval *metav1.Duration `json:"updateInterval"`
}
