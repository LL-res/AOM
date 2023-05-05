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
	})
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
	replica, err := TestScaler.GetModelReplica(predictMetric, 2, UnderThreshold, 3)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(replica)
}
