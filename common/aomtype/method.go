package aomtype

func (h *Hide) Init() {
	h.MetricMap.NewConcurrentMap()
	h.PredictorMap.NewConcurrentMap()
	h.ModelMap.NewConcurrentMap()
	h.CollectorWorkerMap.NewConcurrentMap()
	h.CollectorMap = make(map[string]chan struct{})
}
