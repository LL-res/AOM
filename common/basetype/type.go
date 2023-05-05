package basetype

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Metric struct {
	ScaleDownConf ScaleDownConf `json:"scaleDownConf"`
	Target        float64       `json:"target"`
	Weight        int32         `json:"weight"`
	Name          string        `json:"name"`
	Unit          string        `json:"unit"`
	Query         string        `json:"query"`
}
type ScaleDownConf struct {
	Threshold float64       `json:"threshold"`
	Duration  time.Duration `json:"duration"`
}
type Model struct {
	Type           string           // GRU LSTM
	NeedTrain      bool             //
	UpdateInterval *metav1.Duration `json:"updateInterval"`
	// 移至metric处
	PredcitInterval *metav1.Duration `json:"predcitInterval"`
	// LSTM GRU
	Attr any
}
type LSTM struct {
}
type GRU struct {
	Address        string
	RespRecvAdress string `json:"resp_recv_address"`
	LookBack       int    `json:"look_back"`
	LookForward    int    `json:"look_forward"`
	BatchSize      int    `json:"batch_size"`
	TrainSize      int    `json:"train_size"`
	Epochs         int    `json:"epochs"`
	NLayers        int    `json:"n_layers"`
}
