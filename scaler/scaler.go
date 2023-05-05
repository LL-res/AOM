package scaler

import (
	"errors"
	"github.com/LL-res/AOM/clients/k8s"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

var (
	GlobalScaler *Scaler
)

type Scaler struct {
	namespace      string
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference
	recvChan       chan []float64
}

func (s *Scaler) RecvChan() chan []float64 {
	if s.recvChan == nil {
		s.recvChan = make(chan []float64)
	}
	return s.recvChan
}
func New(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference) *Scaler {

	return &Scaler{namespace: namespace, ScaleTargetRef: scaleTargetRef}

}

// 每个model对应一个
func (s *Scaler) GetModelReplica(predictMetrics []float64, startMetric float64, strategy BaseStrategy, targetMetric float64) ([]int32, error) {
	startReplica, err := k8s.GlobalClient.GetReplica(s.namespace, s.ScaleTargetRef)
	if err != nil {
		return nil, err
	}
	// 获取预测的指标数据
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
	curReplica, err := k8s.GlobalClient.GetReplica(s.namespace, s.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica >= replica {
		return errors.New("target replica num is smaller than the current")
	}
	err = k8s.GlobalClient.SetReplica(s.namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil

}
func (s *Scaler) DownTo(replica int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.namespace, s.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica <= replica {
		return errors.New("target replica num is bigger than the current")
	}
	err = k8s.GlobalClient.SetReplica(s.namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil
}
