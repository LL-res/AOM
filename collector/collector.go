package collector

import (
	"fmt"
	"github.com/LL-res/AOM/common/basetype"

	"sync"
	"time"
)

type MetricSeries struct {
	Metrics   []float64
	TimeStamp []time.Time
	Length    int
}

type WorkerMap[T any] struct {
	data map[string]T
	sync.RWMutex
}

type Collector interface {
	SetServerAddress(url string) error
	ListMetricTypes() []basetype.Metric
	AddCustomMetrics(metric basetype.Metric)
	CreateWorker(MetricType basetype.Metric) (MetricCollector, error)
}
type MetricCollector interface {
	Collect() error
	Send() []Metric
	NoModelKey() string
	DataCap() int
}
type CollectorBase struct {
	//key: the name of  supported metric type,value: the promql to get key metric type
	MetricQL map[basetype.Metric]string
	//server url
	ServerAddress string
}
type MetricType struct {
	Name string
	Unit string
}
type Metric struct {
	Value     float64
	TimeStamp time.Time
}

func (m MetricType) String() string {
	return fmt.Sprintf("%s/%s", m.Name, m.Unit)
}
