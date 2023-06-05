package basetype

type Metric struct {
	ScaleDownConf ScaleDownConf `json:"scaleDownConf"`
	Target        string        `json:"target"`
	// [0,100] each metric weight should have 100 in total
	Weight int32  `json:"weight"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Query  string `json:"query"`
}
type ScaleDownConf struct {
	Threshold string `json:"threshold"`
	Duration  int    `json:"duration"`
}
type Model struct {
	// Type is used to identify the Model and to assert the Attr type
	Type      string `json:"type,omitempty"`
	NeedTrain bool   `json:"needTrain,omitempty"`
	// if NeedTrain is true then UpdateInterval show when to update the model
	UpdateInterval int `json:"updateInterval,omitempty"`
	// e.g. LSTM,GRU
	Attr map[string]string `json:"attr,omitempty"`
}

func (m *Model) DeepCopyInto(out *Model) {
	out.Type = m.Type
	out.NeedTrain = m.NeedTrain
	out.UpdateInterval = m.UpdateInterval
	tmap := make(map[string]string)
	for k, v := range m.Attr {
		tmap[k] = v
	}
}

// model Attr
type LSTM struct {
}

type GRU struct {
	Address        string `json:"address"`
	RespRecvAdress string `json:"resp_recv_address"`
	LookBack       int    `json:"look_back"`
	// all Lookforward should be same
	LookForward int `json:"look_forward"`
	BatchSize   int `json:"batch_size"`
	TrainSize   int `json:"train_size"`
	Epochs      int `json:"epochs"`
	NLayers     int `json:"n_layers"`
}
