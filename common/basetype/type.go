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
	Type            string           // GRU LSTM
	NeedTrain       bool             //
	PredcitInterval *metav1.Duration `json:"predcitInterval"`
	GRU             GRU
	LSTM            LSTM
}
type LSTM struct {
}
type GRU struct {
	// how far in second GRU will use to train
	// +optional
	TrainSize   int    `json:"trainSize"`
	LookBack    int    `json:"lookBack"`
	LookForward int    `json:"lookForward"`
	Address     string `json:"address"`

	//暂时把它当作，需要维持在的值
	ScaleUpThreshold float64 `json:"scaleUpThreshold"`
	//retrain interval
	UpdateInterval *metav1.Duration `json:"updateInterval"`
}
