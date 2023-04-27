package prometheus_collector

import (
	"context"
	"errors"
	"fmt"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"time"
)

type Promc struct {
	collector.CollectorBase
	//prometheus client
	client api.Client
}

func New() collector.Collector {
	metricQL := make(map[basetype.Metric]string, 0)
	cpuAVG := basetype.Metric{
		Name: "avg_node_cpu_usage",
		Unit: "%",
	}
	metricQL[cpuAVG] = "100 - (avg(irate(node_cpu_seconds_total{mode=\"idle\"}[30m])) * 100)"
	res := new(Promc)
	res.MetricQL = metricQL
	return res
}
func (p *Promc) SetServerAddress(url string) error {
	p.ServerAddress = url
	client, err := api.NewClient(api.Config{
		Address: p.ServerAddress,
	})
	if err != nil {
		return err
	}
	p.client = client

	return nil
}

func (p *Promc) ListMetricTypes() []basetype.Metric {
	result := make([]basetype.Metric, 0, len(p.MetricQL))

	for m := range p.MetricQL {
		result = append(result, m)
	}
	return result
}

func (p *Promc) AddCustomMetrics(metricType basetype.Metric) {
	p.MetricQL[metricType] = metricType.Query
}

func (p *Promc) CreateWorker(MetricType basetype.Metric) (collector.MetricCollector, error) {
	promql, ok := p.MetricQL[MetricType]
	if !ok {
		return nil, errors.New("undefined metric type")
	}
	return &worker{
		promql: promql,
		data:   make([]collector.Metric, 0),
		client: p.client,
	}, nil

}

type worker struct {
	collector.MetricType
	promql string
	data   []collector.Metric
	client api.Client
}

func (w *worker) Collect() error {
	v1api := v1.NewAPI(w.client)
	result, _, err := v1api.Query(context.Background(), w.promql, time.Now())
	if err != nil {
		return err
	}
	vector := result.(model.Vector)
	for _, sample := range vector {
		w.data = append(w.data, collector.Metric{
			Value:     float64(sample.Value),
			TimeStamp: sample.Timestamp.Time(),
		})
	}
	return nil
}
func (w *worker) Send() []collector.Metric {
	res := make([]collector.Metric, len(w.data))
	copy(res, w.data)
	w.data = make([]collector.Metric, 0)
	return res
}
func (w *worker) DataCap() int {
	if w.data == nil {
		return 0
	}
	return len(w.data)
}
func (w worker) NoModelKey() string {
	return fmt.Sprintf("%s/%s/%s", w.Name, w.Unit, w.promql)
}
