package predictor

import (
	"context"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
)

type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// predictor is an interface providing methods for making a prediction based on a model, a time to predict and values
type Predictor interface {
	//PredictByReplica(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error)
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
		targetReplicas = append(targetReplicas)
	}
	//策略选择
	return
}

// GetType returns the type of the ModelPredict, "Model"
func (m *ModelPredict) GetType() string {
	return "Models"
}
