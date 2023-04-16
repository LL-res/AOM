package scaler

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/clients/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Scaler struct {
	aom      *automationv1.AOM
	recvChan chan []float64
}

func (s *Scaler) RecvChan() chan []float64 {
	if s.recvChan == nil {
		s.recvChan = make(chan []float64)
	}
	return s.recvChan
}

// 每个model对应一个
func (s *Scaler) GetModelReplica(startMetric float64, strategy BaseStrategy, targetMetric float64) ([]int32, error) {
	scaleObj, err := k8s.GlobalClient.ScaleGetter.Scales(s.aom.Namespace).Get(context.Background(), schema.GroupResource{
		Group:    s.aom.Spec.ScaleTargetRef.APIVersion,
		Resource: s.aom.Spec.ScaleTargetRef.Kind,
	}, s.aom.Spec.ScaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	startReplica := scaleObj.Spec.Replicas
	// 获取预测的指标数据
	predictMetrics := <-s.RecvChan()
	if predictMetrics == nil {
		return nil, errors.New("no metrics received")
	}
	return strategy(targetMetric, startMetric, startReplica, predictMetrics), nil

}

func GetMetricReplica() {

}
func GetObjReplica() {

}
