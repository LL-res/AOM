package scaler

import (
	"fmt"
	"github.com/LL-res/AOM/clients/k8s"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"testing"
)

var TestScaler *Scaler

func TestMain(m *testing.M) {
	TestScaler = New("default", autoscalingv2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       "my-app-deployment",
		APIVersion: "apps/v1",
	}, 5, 3)
	if err := k8s.NewClient(); err != nil {
		panic(err)
	}
	m.Run()
}
func TestScaler_UpTo(t *testing.T) {
	if err := TestScaler.UpTo(2); err != nil {
		t.Error(err)
	}
}
func TestScaler_DownTo(t *testing.T) {
	if err := TestScaler.DownTo(1); err != nil {
		t.Error(err)
	}
}
func TestScaler_GetModelReplica(t *testing.T) {
	predictMetric := make([]float64, 0)
	for i := 0; i < 10; i++ {
		predictMetric = append(predictMetric, float64(i))
	}
	tests := []struct {
		predictMetric []float64
		startMetric   float64
		strategy      BaseStrategy
		targetMetric  float64
	}{
		{
			predictMetric: predictMetric,
			startMetric:   2,
			strategy:      UnderThreshold,
			targetMetric:  3,
		},
		{
			predictMetric: predictMetric,
			startMetric:   2,
			strategy:      Steady,
			targetMetric:  3,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			replica, err := TestScaler.GetModelReplica(test.predictMetric, test.startMetric, test.strategy, test.targetMetric)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Println(replica)
		})
	}

}
func TestScaler_GetMetricReplica(t *testing.T) {
	modelReplica := [][]int32{
		{0, 1, 1, 1, 2, 2, 2, 13, 3, 3},
		{0, 2, 3, 5, 6, 8, 9, 11, 12, 14},
	}
	tests := []struct {
		modelReplica [][]int32
		strategy     ModelStrategy
	}{
		{
			modelReplica: modelReplica,
			strategy:     MaxStrategy,
		},
		{
			modelReplica: modelReplica,
			strategy:     MinStrategy,
		},
		{
			modelReplica: modelReplica,
			strategy:     MeanStrategy,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			metricReplica := TestScaler.GetMetricReplica(test.modelReplica, test.strategy)
			fmt.Println(metricReplica)
		})
	}
}

func TestScaler_GetScaleReplica(t *testing.T) {
	tests := []struct {
		ObjReplica []int32
		strategy   ObjStrategy
	}{
		{
			ObjReplica: []int32{0, 1, 1, 1, 2, 2, 2, 11, 3, 3},
			strategy:   SelectMax,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			replica := TestScaler.GetScaleReplica(test.ObjReplica, test.strategy)
			fmt.Println(replica)
		})
	}
}
