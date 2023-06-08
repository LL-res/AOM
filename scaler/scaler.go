package scaler

import (
	"errors"
	"fmt"
	"github.com/LL-res/AOM/clients/k8s"
	"github.com/LL-res/AOM/log"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type Scaler struct {
	MaxReplica     int32  `json:"maxReplica"`
	MinReplica     int32  `json:"minReplica"`
	Namespace      string `json:"namespace"`
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference
	recvChan       chan []float64
}

func (s *Scaler) RecvChan() chan []float64 {
	if s.recvChan == nil {
		s.recvChan = make(chan []float64)
	}
	return s.recvChan
}
func (s *Scaler) New(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference, maxReplica, minReplica int32) *Scaler {
	return &Scaler{MinReplica: minReplica, MaxReplica: maxReplica, Namespace: namespace, ScaleTargetRef: scaleTargetRef}
}
func New(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference, maxReplica, minReplica int32) *Scaler {

	return &Scaler{MinReplica: minReplica, MaxReplica: maxReplica, Namespace: namespace, ScaleTargetRef: scaleTargetRef}

}

// 每个model对应一个
func (s *Scaler) GetModelReplica(predictMetrics []float64, startMetric float64, strategy BaseStrategy, targetMetric float64) ([]int32, error) {
	startReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
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
	curReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica >= replica {
		log.Logger.Info("do not scale", "scale target", s.ScaleTargetRef, "current replica", fmt.Sprint(curReplica), "target replica", fmt.Sprint(replica))
		return errors.New("target replica num is smaller than the current")
	}
	if replica > s.MaxReplica {
		log.Logger.Info("do not scale", "scale target", s.ScaleTargetRef, "max replica", fmt.Sprint(s.MaxReplica), "target replica", fmt.Sprint(replica))
		return nil
	}
	err = k8s.GlobalClient.SetReplica(s.Namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil

}
func (s *Scaler) DownTo(replica int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica <= replica {
		return errors.New("target replica num is bigger than the current")
	}
	err = k8s.GlobalClient.SetReplica(s.Namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil
}
