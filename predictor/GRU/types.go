package GRU

import (
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/predictor"
	"go.uber.org/atomic"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type GRU struct {
	predictor.Base
	withModelKey    string
	model           automationv1.Model
	collectorWorker collector.MetricCollector
	readyToPredict  *atomic.Bool
	address         string
	ScaleTargetRef  autoscalingv2.CrossVersionObjectReference
}

type Request struct {
	PredictHistory []float64 `json:"predict_history"`
	TrainHistory   []float64 `json:"train_history"`
	RespRecvAdress string    `json:"resp_recv_address"`
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
