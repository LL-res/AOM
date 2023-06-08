package holt_winter

import (
	"context"
	"encoding/json"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/common/errs"
	ptype "github.com/LL-res/AOM/predictor/type"
	"strconv"
)

type HoltWinter struct {
	slen            int
	lookForward     int
	lookBackward    int
	alpha           float64
	beta            float64
	gamma           float64
	withModelKey    string
	collectorWorker collector.MetricCollector
}
type Param struct {
	Slen         string `json:"slen,omitempty"`
	LookForward  string `json:"look_forward,omitempty"`
	LookBackward string `json:"look_backward,omitempty"`
	Alpha        string `json:"alpha,omitempty"`
	Beta         string `json:"beta,omitempty"`
	Gamma        string `json:"gamma,omitempty"`
}

func (p *HoltWinter) Predict(ctx context.Context) (ptype.PredictResult, error) {
	if p.collectorWorker.DataCap() < p.lookBackward {
		return ptype.PredictResult{}, errs.NO_SUFFICENT_DATA
	}
	metrcis := p.collectorWorker.Send()
	metrcis = metrcis[len(metrcis)-p.lookBackward:]
	series := make([]float64, 0)
	for _, m := range metrcis {
		series = append(series, m.Value)
	}
	predMetrics := p.tripleExponentialSmoothing(series)
	res := ptype.PredictResult{
		StartMetric:   series[len(series)-1],
		Loss:          -1,
		PredictMetric: predMetrics,
	}
	return res, nil
}

func (p *HoltWinter) GetType() string {
	return consts.HOLT_WINTER
}

func (p *HoltWinter) Train(ctx context.Context) error {
	return nil
}

func (p *HoltWinter) Key() string {
	return p.withModelKey
}

func (p *HoltWinter) initialTrend(series []float64) float64 {
	sum := 0.0
	for i := 0; i < p.slen; i++ {
		sum += (series[i+p.slen] - series[i]) / float64(p.slen)
	}
	return sum / float64(p.slen)
}

func (p *HoltWinter) initialSeasonalComponents(series []float64) map[int]float64 {
	seasonals := make(map[int]float64)
	seasonAverages := make([]float64, 0)
	nSeasons := len(series) / p.slen

	// Compute season averages
	for j := 0; j < nSeasons; j++ {
		sum := 0.0
		for _, value := range series[p.slen*j : p.slen*j+p.slen] {
			sum += value
		}
		seasonAverages = append(seasonAverages, sum/float64(p.slen))
	}

	// Compute initial values
	for i := 0; i < p.slen; i++ {
		sumOfValsOverAvg := 0.0
		for j := 0; j < nSeasons; j++ {
			sumOfValsOverAvg += series[p.slen*j+i] - seasonAverages[j]
		}
		seasonals[i] = sumOfValsOverAvg / float64(nSeasons)
	}

	return seasonals
}

func (p *HoltWinter) tripleExponentialSmoothing(series []float64) []float64 {
	result := make([]float64, 0)
	seasonals := p.initialSeasonalComponents(series)
	smooth := series[0]
	trend := p.initialTrend(series)

	for i := 0; i < len(series)+p.lookForward; i++ {
		if i == 0 {
			result = append(result, series[0])
			continue
		}
		if i >= len(series) {
			m := i - len(series) + 1
			result = append(result, (smooth+float64(m)*trend)+seasonals[i%p.slen])
		} else {
			val := series[i]
			lastSmooth := smooth
			smooth = p.alpha*(val-seasonals[i%p.slen]) + (1-p.alpha)*(smooth+trend)
			trend = p.beta*(smooth-lastSmooth) + (1-p.beta)*trend
			seasonals[i%p.slen] = p.gamma*(val-smooth) + (1-p.gamma)*seasonals[i%p.slen]
			result = append(result, smooth+trend+seasonals[i%p.slen])
		}
	}

	return result
}
func New(collectorWorker collector.MetricCollector, model map[string]string, withModelKey string) (*HoltWinter, error) {
	j, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	param := Param{}
	err = json.Unmarshal(j, &param)
	if err != nil {
		return nil, err
	}
	slen, err := strconv.Atoi(param.Slen)
	if err != nil {
		return nil, err
	}
	lookForward, err := strconv.Atoi(param.LookForward)
	if err != nil {
		return nil, err
	}
	lookBack, err := strconv.Atoi(param.LookBackward)
	if err != nil {
		return nil, err
	}
	alpha, err := strconv.ParseFloat(param.Alpha, 64)
	if err != nil {
		return nil, err
	}
	beta, err := strconv.ParseFloat(param.Beta, 64)
	if err != nil {
		return nil, err
	}
	gamma, err := strconv.ParseFloat(param.Gamma, 64)
	if err != nil {
		return nil, err
	}
	return &HoltWinter{
		slen:            slen,
		lookForward:     lookForward,
		lookBackward:    lookBack,
		alpha:           alpha,
		beta:            beta,
		gamma:           gamma,
		withModelKey:    withModelKey,
		collectorWorker: collectorWorker,
	}, nil

}
