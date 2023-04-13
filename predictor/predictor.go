package predictor

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor/GRU"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}
type Param struct {
	ModelType       string
	MetricCollector collector.MetricCollector
	Model           automationv1.Model
	Adress          string
	ScaleTargetRef  autoscalingv2.CrossVersionObjectReference
}

// predictor is an interface providing methods for making a prediction based on a model, a time to predict and values
type Predictor interface {
	Predict(ctx context.Context, aom *automationv1.AOM) ([]int32, error)
	GetType() string
	Train(ctx context.Context) error
}
type Base struct {
	MetricHistory []collector.Metric // 存储着全部
	//socket client
}

// 多模型，根据一定的策略进行预测结果的选取
type ModelPredict struct {
	predictors []Predictor
}

func (m *ModelPredict) Predict(ctx context.Context, aom *automationv1.AOM) ([]int32, error) {
	//此处存放着所有模型预测出的结果
	targetReplicas := make([][]int32, 0)
	for _, predictor := range m.predictors {
		res, err := predictor.Predict(ctx, aom)
		if err != nil {
			return nil, err
		}
		targetReplicas = append(targetReplicas, res)
	}
	//策略选择

	return nil, nil
}

// GetType returns the type of the ModelPredict, "Model"
func (m *ModelPredict) GetType() string {
	return "Models"
}
func (m *ModelPredict) Train(ctx context.Context) error {
	return nil
}
func NewPredictor(param Param) (Predictor, error) {
	switch param.ModelType {
	case automationv1.TypeGRU:
		return GRU.New(param.MetricCollector, param.Model, param.Adress, param.ScaleTargetRef)
	default:
		return nil, errors.New("unknown predictor")
	}
}
