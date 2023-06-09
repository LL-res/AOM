package predictor

import (
	"context"
	"errors"
	"github.com/LL-res/AOM/algorithms/holt_winter"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/predictor/GRU"
	ptype "github.com/LL-res/AOM/predictor/type"
	"github.com/LL-res/AOM/utils"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type Param struct {
	WithModelKey    string
	MetricCollector collector.MetricCollector
	Model           map[string]string
	ScaleTargetRef  autoscalingv2.CrossVersionObjectReference
}

// predictor is an interface providing methods for making a prediction based on a model, a time to predict and values
type Predictor interface {
	Predict(ctx context.Context) (ptype.PredictResult, error)
	GetType() string
	Train(ctx context.Context) error
	Key() string
}
type Base struct {
	MetricHistory []collector.Metric // 存储着全部
	//socket client
}

// // 多模型，根据一定的策略进行预测结果的选取
//
//	type ModelPredict struct {
//		predictors []Predictor
//	}
type PredictResult struct {
	StartMetric   float64
	StartReplica  int32
	PredictMetric []float64
}

//	func (m *ModelPredict) Predict(ctx context.Context, aom *automationv1.AOM) (PredictResult, error) {
//		//此处存放着所有模型预测出的结果
//		targetReplicas := make([][]float64, 0)
//		for _, predictor := range m.predictors {
//			res, err := predictor.Predict(ctx, aom)
//			if err != nil {
//				return nil, err
//			}
//			targetReplicas = append(targetReplicas, res)
//		}
//		//策略选择
//
//		return nil, nil
//	}
//
// // GetType returns the type of the ModelPredict, "Model"
//
//	func (m *ModelPredict) GetType() string {
//		return "Models for one metric"
//	}
//
//	func (m *ModelPredict) Train(ctx context.Context) error {
//		return nil
//	}
func NewPredictor(param Param) (Predictor, error) {
	switch utils.GetModelType(param.WithModelKey) {
	case consts.GRU:
		pred, err := GRU.New(param.MetricCollector, param.Model, param.WithModelKey)
		if err != nil {
			return nil, err
		}
		return pred, nil
	case consts.HOLT_WINTER:
		pred, err := holt_winter.New(param.MetricCollector, param.Model, param.WithModelKey)
		if err != nil {
			return nil, err
		}
		return pred, nil
	default:
		return nil, errors.New("unknown predictor")
	}
}
