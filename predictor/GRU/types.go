package GRU

import (
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/basetype"
	ptype "github.com/LL-res/AOM/predictor/type"
	"go.uber.org/atomic"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type GRU struct {
	ptype.Base
	withModelKey    string
	model           basetype.GRU
	collectorWorker collector.MetricCollector
	readyToPredict  *atomic.Bool
	address         string
	ScaleTargetRef  autoscalingv2.CrossVersionObjectReference
	debug           bool
}

type Request struct {
	Key            string    `json:"key"`
	PredictHistory []float64 `json:"predict_history,omitempty"`
	TrainHistory   []float64 `json:"train_history,omitempty"`
	RespRecvAdress string    `json:"resp_recv_address"`
	LookBack       int       `json:"look_back"`
	LookForward    int       `json:"look_forward"`
	BatchSize      int       `json:"batch_size,omitempty"`
	Epochs         int       `json:"epochs,omitempty"`
	NLayers        int       `json:"n_layers,omitempty"`
}

type Response struct {
	//模型训练的误差参数，例如均方误差值
	Loss       float64   `json:"loss"`
	Trained    bool      `json:"trained"`
	Prediction []float64 `json:"prediction"`
	Error      string    `json:"error"`
}

// spec fields
type Param struct {
	Address        string `json:"address"`
	RespRecvAdress string `json:"resp_recv_address"`
	LookBack       string `json:"look_back"`
	LookForward    string `json:"look_forward"`
	BatchSize      string `json:"batch_size"`
	TrainSize      string `json:"train_size"`
	Epochs         string `json:"epochs"`
	NLayers        string `json:"n_layers"`
	Debug          string `json:"debug"`
}
