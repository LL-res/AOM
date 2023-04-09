package predictor

import (
	"fmt"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
)

type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// predictor is an interface providing methods for making a prediction based on a model, a time to predict and values
type Predictor interface {
	//PredictByReplica(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error)
	Predict(model *automationv1.Model, metricHistory []collector.Metric) (int32, error)
	PruneHistory(model *automationv1.Model, replicaHistory []collector.Metric) ([]collector.Metric, error)
	GetType() string
	Train() error
}
type Base struct {
	MetricHistory []collector.Metric // 存储着全部
	//socket client
}

// ModelPredict is used to route a prediction to the appropriate predictor based on the model provided
// Should be initialised with available predictors for it to use
type ModelPredict struct {
	predictors []Predictor
}

// GetPrediction generates a prediction for any model that the ModelPredict has been set up to use
func (m *ModelPredict) Predict(model *automationv1.Model, metricHistory []collector.Metric) (int32, error) {
	for _, predictor := range m.predictors {
		if predictor.GetType() == model.Type {
			return predictor.Predict(model, metricHistory)
		}
	}
	return 0, fmt.Errorf("unknown model type '%s'", model.Type)
}

// GetIDsToRemove finds the appropriate logic for the model and gets a list of stored IDs to remove
func (m *ModelPredict) PruneHistory(model *automationv1.Model, replicaHistory []collector.Metric) ([]collector.Metric, error) {
	for _, predictor := range m.predictors {
		if predictor.GetType() == model.Type {
			return predictor.PruneHistory(model, replicaHistory)
		}
	}
	return nil, fmt.Errorf("unknown model type '%s'", model.Type)
}

// GetType returns the type of the ModelPredict, "Model"
func (m *ModelPredict) GetType() string {
	return "Models"
}
