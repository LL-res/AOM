package GRU

import (
	"context"
	"fmt"
	"github.com/LL-res/AOM/clients/k8s"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/fake"
	"os"
	"os/exec"
	"testing"
	"time"
)

var (
	ctx context.Context
	gru *GRU
)

func TestGRU_Predict(t *testing.T) {
	result, err := gru.Predict(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(result)
}
func TestGRU_Train(t *testing.T) {
	err := gru.Train(ctx)
	if err != nil {
		t.Error(err)
	}
	select {}
}
func TestMain(m *testing.M) {
	if err := k8s.NewClient(); err != nil {
		panic(err)
	}
	ctx = context.WithValue(context.Background(), consts.NAMESPACE, "default")
	//collectWorker := fake.CollectorWorker{
	//	N: 1000,
	//	Function: func(i int) float64 {
	//		return math.Sin(float64(i) * 0.01)
	//	},
	//	Start:    time.Now(),
	//	Interval: 5 * time.Second,
	//}
	//
	//gruAttr := map[string]string{
	//	"address":           "/tmp/uds_socket",
	//	"resp_recv_address": RespRecvAdress,
	//	"look_back":         "100",
	//	"look_forward":      "60",
	//	"train_size":        "1000",
	//	"epochs":            "1",
	//	"n_layers":          "2",
	//}
	//
	//var err error
	//gru, err = New(&collectWorker, gruAttr, collectWorker.NoModelKey()+"$gru")
	//if err != nil {
	//	panic(err)
	//}
	m.Run()
}
func TestPython(t *testing.T) {
	fmt.Println(os.Getwd())
	cmd := exec.Command(PYTHON, pythonDir)
	if err := cmd.Run(); err != nil {
		t.Error(err)
	}
}
func TestProcess(t *testing.T) {
	tests := []struct {
		collector collector.MetricCollector
		attr      map[string]string
	}{
		{
			collector: &fake.CollectorWorker{
				N: 1000,
				Function: func(i int) float64 {
					base := []int{
						30, 21, 29, 31, 40, 48, 53, 47, 37, 39, 31, 29, 17, 9, 20, 24, 27, 35, 41, 38,
						27, 31, 27, 26, 21, 13, 21, 18, 33, 35, 40, 36, 22, 24, 21, 20, 17, 14, 17, 19,
						26, 29, 40, 31, 20, 24, 18, 26, 17, 9, 17, 21, 28, 32, 46, 33, 23, 28, 22, 27,
						18, 8, 17, 21, 31, 34, 44, 38, 31, 30, 26, 32,
					}
					return float64(base[i%len(base)])
				},
				Start:    time.Now(),
				Interval: 5 * time.Second,
			},
			attr: map[string]string{
				"address":           "/tmp/uds_socket",
				"resp_recv_address": RespRecvAdress,
				"look_back":         "100",
				"look_forward":      "60",
				"train_size":        "1000",
				"batch_size":        "100",
				"epochs":            "100",
				"n_layers":          "2",
				"debug":             "true",
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			localGRU, err := New(tt.collector, tt.attr, tt.collector.NoModelKey()+"$gru")
			if err != nil {
				t.Error(err)
				return
			}
			err = localGRU.Train(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			for !localGRU.readyToPredict.Load() {

			}
			res, err := localGRU.Predict(ctx)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Println(res)
		})
	}

}
