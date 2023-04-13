package collector

import (
	"fmt"
	"sync"
	"time"
)

type MetricSeries struct {
	Metrics   []float64
	TimeStamp []time.Time
	Length    int
}

var GlobalMetricCollectorMap *WorkerMap

type WorkerMap struct {
	data map[string]MetricCollector
	sync.RWMutex
}

func InitGlobalMap() {
	GlobalMetricCollectorMap = &WorkerMap{
		data: make(map[string]MetricCollector),
	}
}

func (m *WorkerMap) Load(noModelKey string) (MetricCollector, bool) {
	m.RLock()
	defer m.RUnlock()
	worker, ok := m.data[noModelKey]
	return worker, ok
}
func (m *WorkerMap) Store(noModelKey string, worker MetricCollector) {
	m.Lock()
	defer m.Unlock()
	m.data[noModelKey] = worker
}
func (m *WorkerMap) Delete(noModelKey string) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, noModelKey)
}

type Collector interface {
	SetServerAddress(url string) error
	ListMetricTypes() []MetricType
	AddCustomMetrics(metric MetricType, query string)
	CreateWorker(MetricType MetricType) (MetricCollector, error)
}
type MetricCollector interface {
	Collect() error
	Send() []Metric
	NoModelKey() string
	DataCap() int
}
type CollectorBase struct {
	//key: the name of  supported metric type,value: the promql to get key metric type
	MetricQL map[MetricType]string
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
