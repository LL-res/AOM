package fake

import (
	"github.com/LL-res/AOM/collector"
	"time"
)

type CollectorWorker struct {
	// 生成的点的数量
	N int
	// 生成的点的函数走势
	Function func(i int) float64
	// 生成点的开始时间
	Start time.Time
	// 生成点的时间间隔
	Interval time.Duration
}

func (c *CollectorWorker) Send() []collector.Metric {
	res := make([]collector.Metric, 0)
	for i := 0; i < c.N; i++ {
		res = append(res, collector.Metric{
			Value:     c.Function(i),
			TimeStamp: c.Start.Add(time.Duration(i) * c.Interval),
		})
	}
	return res
}

func (c *CollectorWorker) NoModelKey() string {
	return "fake$%$test_fake"
}

func (c *CollectorWorker) DataCap() int {
	return 1<<32 - 1
}

func (c *CollectorWorker) Collect() error {
	return nil
}
