package basetype

import (
	"time"
)

type Metric struct {
	ScaleDownConf ScaleDownConf `json:"scaleDownConf"`
	Target        float64       `json:"target"`
	// [0,100] each metric weight should have 100 in total
	Weight int32  `json:"weight"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Query  string `json:"query"`
}
type ScaleDownConf struct {
	Threshold float64       `json:"threshold"`
	Duration  time.Duration `json:"duration"`
}
type Model struct {
	// Type is used to identify the Model and to assert the Attr type
	Type      string
	NeedTrain bool
	// if NeedTrain is true then UpdateInterval show when to update the model
	UpdateInterval time.Duration `json:"updateInterval"`
	// e.g. LSTM,GRU
	Attr any
}

// model Attr
type LSTM struct {
}

type GRU struct {
	Address        string
	RespRecvAdress string `json:"resp_recv_address"`
	LookBack       int    `json:"look_back"`
	// all Lookforward should be same
	LookForward int `json:"look_forward"`
	BatchSize   int `json:"batch_size"`
	TrainSize   int `json:"train_size"`
	Epochs      int `json:"epochs"`
	NLayers     int `json:"n_layers"`
}
