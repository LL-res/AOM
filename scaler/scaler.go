package scaler

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/clients/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GlobalScaler *Scaler
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
func Init(aom *automationv1.AOM) {
	if GlobalScaler == nil {
		GlobalScaler = &Scaler{aom: aom}
	}
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

// 获取每个metric对应的预测样本数
func (s *Scaler) GetMetricReplica(modelReplica [][]int32, strategy ModelStrategy) []int32 {
	return strategy(modelReplica)
}

// 获取到了被检测对象之后时间端的样本数
// 之后的操作应该是从这个切片中进行选取，选取一个或是多个合适的值，作为在当前时刻要进行的扩缩容副本数
func (s *Scaler) GetObjReplica(metricReplica [][]int32, strategy MetricStrategy) []int32 {
	return strategy(metricReplica)
}

func (s *Scaler) GetScaleReplica(objReplicaSet []int32, strategy ObjStrategy) int32 {
	return strategy(objReplicaSet)
}
func (s *Scaler) UpTo(replica int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.aom.Namespace, s.aom.Spec.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica >= replica {
		return errors.New("target replica num is smaller than the current")
	}
	err = k8s.GlobalClient.SetReplica(s.aom.Namespace, s.aom.Spec.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil

}
func (s *Scaler) DownTo(replica int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.aom.Namespace, s.aom.Spec.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica <= replica {
		return errors.New("target replica num is bigger than the current")
	}
	err = k8s.GlobalClient.SetReplica(s.aom.Namespace, s.aom.Spec.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil
}
