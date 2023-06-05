/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/LL-res/AOM/common/basetype"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AOMSpec defines the desired state of AOM
type AOMSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference `json:"scaleTargetRef"`

	// +kubebuilder:validation:Minimum=0
	// +optional
	MinReplicas *int32 `json:"minReplicas"`

	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas int32                       `json:"maxReplicas"`
	Collector   Collector                   `json:"collector"`
	Metrics     map[string]basetype.Metric  `json:"metrics"`
	Models      map[string][]basetype.Model `json:"models"`
	// the interval aom to call all the model
	Interval int `json:"interval"`
}
type Collector struct {
	Address string `json:"address"`
	// LookForward * ScrapInterval = the time to look forward
	ScrapeInterval int `json:"scrapeInterval"`
}

//type Metric struct {
//	ScaleDownConf ScaleDownConf `json:"scaleDownConf"`
//	Target        float64       `json:"target"`
//	Name          string        `json:"name"`
//	Unit          string        `json:"unit"`
//	Query         string        `json:"query"`
//}
//type ScaleDownConf struct {
//	Threshold float64       `json:"threshold"`
//	Duration  time.Duration `json:"duration"`
//}

//type Model struct {
//	Type            string           // GRU LSTM
//	PredcitInterval *metav1.Duration `json:"predcitInterval"`
//	GRU             GRU
//	LSTM            LSTM
//}
//type LSTM struct {
//}
//type GRU struct {
//	// how far in second GRU will use to train
//	// +optional
//	TrainSize   int    `json:"trainSize"`
//	LookBack    int    `json:"lookBack"`
//	LookForward int    `json:"lookForward"`
//	Address     string `json:"address"`
//
//	//暂时把它当作，需要维持在的值
//	ScaleUpThreshold float64 `json:"scaleUpThreshold"`
//	//retrain interval
//	UpdateInterval *metav1.Duration `json:"updateInterval"`
//}

//type PredictorHistory struct {
//	PredictHistory []time.Time
//	TrainHistory   []time.Time
//}
//
//func appendHistory(history *[]time.Time, t time.Time) {
//	if history == nil {
//		*history = make([]time.Time, 0, 5)
//	}
//
//	*history = append(*history, t)
//
//	if len(*history) > 5 {
//		*history = (*history)[1:]
//	}
//	return
//}
//
//func (p *PredictorHistory) CanTrain(now time.Time, interval time.Duration) bool {
//	if p.TrainHistory == nil || len(p.TrainHistory) == 0 {
//		return true
//	}
//	return p.TrainHistory[len(p.TrainHistory)-1].Add(interval).Before(now)
//}
//
//func (p *PredictorHistory) CanPredict(now time.Time, interval time.Duration) bool {
//	if p.PredictHistory == nil || len(p.PredictHistory) == 0 {
//		return true
//	}
//	return p.PredictHistory[len(p.PredictHistory)-1].Add(interval).Before(now)
//}
//
//func (p *PredictorHistory) AppendPredictorHistory(t time.Time) {
//	appendHistory(&p.PredictHistory, t)
//}
//
//func (p *PredictorHistory) AppendTrainHistory(t time.Time) {
//	appendHistory(&p.TrainHistory, t)
//}

// AOMStatus defines the observed state of AOM
type AOMStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// up or down
	// noModelKey
	//CollectorMap map[string]chan struct{} `json:"-"`
	// withModelKey
	//PredictorMap     map[string]struct{} `json:"-"`
	StatusCollectors []StatusCollector `json:"collectors"`
	// withModelKey
	//PredictorHistory utils.ConcurrentMap[*PredictorHistory] `json:"-"`
	Generation int64 `json:"generation"`
}
type StatusCollector struct {
	Name       string `json:"name,omitempty"`
	Unit       string `json:"unit,omitempty"`
	Expression string `json:"expression,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AOM is the Schema for the aoms API
type AOM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AOMSpec   `json:"spec,omitempty"`
	Status AOMStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AOMList contains a list of AOM
type AOMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AOM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AOM{}, &AOMList{})
}
