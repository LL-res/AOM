package GRU

import (
	"encoding/json"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor"
	"net"

	"log"
	"sort"
	"time"
)

const (
	defaultTimeout = 30000
	algorithmPath  = "algorithms/linear_regression/linear_regression.py"
)

type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// 在controller里控制要不要进行predict或者train，status里记录了模型的状态
// 在这里的预测服务只要专心进行预测即可
// Predict provides logic for using GRU to make a prediction
type GRU struct {
	predictor.Base
	model           automationv1.Model
	collectorWorker collector.MetricCollector
	readyToPredict  bool
	address         string
}
type GRUParameters struct {
	LookAhead      time.Duration                                `json:"lookAhead"`
	MetricsHistory []jamiethompsonmev1alpha1.TimestampedMetrics `json:"metricsHistory"`
	// predict or train
	Status int `json:"status"`
}
type GRUResult struct {
	Value   int  `json:"value"`
	Trained bool `json:"trained"`
}
type Request struct {
	PredictHistory []float64 `json:"predict_history"`
	TrainHistory   []float64 `json:"train_history"`
	LookBack       int       `json:"look_back"`
	LookForward    int       `json:"look_forward"`
	BatchSize      int       `json:"batch_size"`
	Epochs         int       `json:"epochs"`
	NLayers        int       `json:"n_layers"`
}
type Response struct {
	Trained    bool      `json:"trained"`
	Prediction []float64 `json:"prediction"`
}

func New(collectorWorker collector.MetricCollector, model automationv1.Model, address string) (*GRU, error) {
	return &GRU{
		Base: predictor.Base{
			MetricHistory: collectorWorker.Send(),
		},
		model:           model,
		collectorWorker: collectorWorker,
		readyToPredict:  false,
		address:         address,
	}, nil

}
func (g *GRU) Predict(lastUpdateTime time.Time) (int32, error) {
	if !g.readyToPredict {
		return 0, errors.New("the model is not ready to predict")
	}
	// 如果worker中的数据量不足，直接返回
	if g.collectorWorker.DataCap() < g.model.GRU.LookForward {
		return 0, errors.New("no sufficient data to predict")
	}
	tempData := g.collectorWorker.Send()
	// with timestamp
	predictData := tempData[len(tempData)-g.model.GRU.LookBack:]
	//no timestamp
	predictHistory := make([]float64, 0, len(predictData))
	for _, v := range predictData {
		predictHistory = append(predictHistory, v.Value)
	}
	req := Request{
		PredictHistory: predictHistory,
		LookBack:       g.model.GRU.LookBack,
		LookForward:    g.model.GRU.LookForward,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	conn, err := net.Dial("unix", g.address)
	if err != nil {
		return 0, err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	// 客户端给
	_, err = conn.Write(append(reqJson, '\n'))

	if err != nil {
		return 0, err
	}

	conn.Read()

	if model.GRU == nil {
		return 0, errors.New("no GRU configuration provided for model")
	}
	if !p.trained {
		return 0, errors.New("model not trained")
	}
	var status int
	now := time.Now()
	// 00 stand for train predict
	//e.g. 01 predict ,11 train and predict,10 train
	// predictSize < trainSize
	switch {
	case len(metricsHistory) < model.GRU.PredictSize:
		return 0, errors.New("no sufficient data to train or predict")
	case len(metricsHistory) > model.GRU.TrainSize && now.Add(-model.GRU.UpdateInterval.Duration).After(lastUpdateTime):
		status = 3
	default:
		status = 1
	}
	parameters, err := json.Marshal(GRUParameters{
		LookAhead:      model.GRU.LookAhead,
		MetricsHistory: metricsHistory,
		Status:         status,
	})
	if err != nil {
		panic(err)
	}
	timeout := defaultTimeout
	if model.CalculationTimeout != nil {
		timeout = *model.CalculationTimeout
	}

	data, err := p.Runner.RunAlgorithmWithValue(algorithmPath, string(parameters), timeout)
	if err != nil {
		return 0, err
	}
	res := GRUResult{}
	err = json.Unmarshal([]byte(data), &res)
	if err != nil {
		return 0, err
	}
	p.trained = res.Trained
	return int32(res.Value), nil
}

func (g *GRU) PruneHistory(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
	if model.Linear == nil {
		return nil, errors.New("no GRU configuration provided for model")
	}

	if len(replicaHistory) < model.Linear.HistorySize {
		return replicaHistory, nil
	}

	// Sort by date created, newest first
	sort.Slice(replicaHistory, func(i, j int) bool {
		return !replicaHistory[i].Time.Before(replicaHistory[j].Time)
	})

	// Remove oldest to fit into requirements, have to loop from the end to allow deletion without affecting indices
	for i := len(replicaHistory) - 1; i >= model.Linear.HistorySize; i-- {
		replicaHistory = append(replicaHistory[:i], replicaHistory[i+1:]...)
	}

	return replicaHistory, nil
}

func (g *GRU) GetType() string {
	return jamiethompsonmev1alpha1.TypeGRU
}
func (g *GRU) Train(model *jamiethompsonmev1alpha1.Model) error {
	if len(g.MetricHistory) < model.GRU.TrainSize {
		return errors.New("no sufficient data to train or predict")
	}
	parameters, err := json.Marshal(GRUParameters{
		LookAhead:      model.GRU.LookAhead,
		MetricsHistory: g.MetricHistory,
	})
	if err != nil {
		return err
	}
	timeout := defaultTimeout
	res, err := g.Runner.RunAlgorithmWithValue(algorithmPath, string(parameters), timeout)
	if err != nil {
		log.Println(err)
		return err
	}

}
