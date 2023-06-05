package holt_winter

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/common/consts"
	ptype "github.com/LL-res/AOM/predictor/type"
)

type HoltWinter struct {
	slen            int
	lookForward     int
	lookBack        int
	alpha           float64
	beta            float64
	gamma           float64
	withModelKey    string
	collectorWorker collector.MetricCollector
}
type Param struct {
	Slen        int     `json:"slen,omitempty"`
	LookForward int     `json:"look_forward,omitempty"`
	LookBack    int     `json:"look_back,omitempty"`
	Alpha       float64 `json:"alpha,omitempty"`
	Beta        float64 `json:"beta,omitempty"`
	Gamma       float64 `json:"gamma,omitempty"`
}

func (p *HoltWinter) Predict(ctx context.Context) (ptype.PredictResult, error) {
	if p.collectorWorker.DataCap() < p.lookBack {
		return ptype.PredictResult{}, errors.New("no sufficient data to predict")
	}
	metrcis := p.collectorWorker.Send()
	metrcis = metrcis[len(metrcis)-p.lookBack:]
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
	return &HoltWinter{
		slen:            param.Slen,
		lookForward:     param.LookForward,
		lookBack:        param.LookBack,
		alpha:           param.Alpha,
		beta:            param.Beta,
		gamma:           param.Gamma,
		withModelKey:    withModelKey,
		collectorWorker: collectorWorker,
	}, nil

}
