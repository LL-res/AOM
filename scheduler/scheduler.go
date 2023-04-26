package scheduler

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/scaler"
	"log"
	"time"
)

var (
	GlobalScheduler *Scheduler
)

type Scheduler struct {
	aom        *automationv1.AOM
	predictors map[automationv1.Metric][]predictor.Predictor
}
type Conf struct {
	needTrain       bool
	trainInterval   time.Duration
	predictInterval time.Duration
}

func Init(aom *automationv1.AOM) {
	if GlobalScheduler == nil {
		GlobalScheduler = &Scheduler{aom: aom}
	}
}

// predictors : 每个metric指标对应一组predictor，predictors中包含一个aom实例所拥有的全部的predictor，并按所属metric不同，分为不同的组
func (s *Scheduler) HandlePredictors(ctx context.Context, predictors map[automationv1.Metric][]predictor.Predictor) error {
	// 何时预测，何时更新的信息记录在model中
	// 生成一个map，记录predictor与model的映射关系
	// status 里应该记录历史数据，与时间戳，以检查是否需要进行操作
	metricReplicas := make([][]int32, 0)
	for m, p := range predictors {
		metricReplica, err := s.HandleForOneMetric(ctx, m, p)
		if err != nil {
			log.Println(err)
			continue
		}
		metricReplicas = append(metricReplicas, metricReplica)
	}
	objSet := scaler.GlobalScaler.GetObjReplica(metricReplicas, scaler.MaxStrategy)
	targetReplica := scaler.GlobalScaler.GetScaleReplica(objSet, scaler.SelectMax)

	if err := scaler.GlobalScaler.UpTo(targetReplica); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Scheduler) HandleForOneMetric(ctx context.Context, metric automationv1.Metric, predictors []predictor.Predictor) ([]int32, error) {

	metricReplica := make([]int32, 0)

	for _, p := range predictors {
		history, ok := s.aom.Status.PredictorHistory.Load(p.Key())
		if !ok {
			// 此刻status中还没有任何history存在说明此刻predictor还未进行过动作,则应进行初始化
			s.aom.Status.PredictorHistory.Store(p.Key(), new(automationv1.PredictorHistory))
		}
		model, ok := predictor.PredictorModelMap.Load(p.Key())
		if !ok {
			return nil, errors.New("model not find")
		}
		//拿到了model与history，history中包含了历史记录
		//model中包含着对predictor进行调用的时间戳
		conf := Conf{
			needTrain:       false,
			predictInterval: model.PredcitInterval.Duration,
		}
		switch model.Type {
		case automationv1.TypeGRU, automationv1.TypeLSTM:
			conf.trainInterval = model.GRU.UpdateInterval.Duration
			conf.needTrain = true
		}
		now := time.Now()

		modelReplicas := make([][]int32, 0)
		if history.CanPredict(now, conf.predictInterval) {
			var err error
			pResult, err := p.Predict(ctx, s.aom)
			if err != nil { // 此时的错误包含了数据不足，模型未训练等
				log.Println(err)
				continue
				//此时仅对err进行打印
			} else {
				temp, err := scaler.GlobalScaler.GetModelReplica(pResult.StartMetric, scaler.Steady, metric.Target)
				if err != nil {
					log.Println(err)
					continue
				}
				modelReplicas = append(modelReplicas, temp)
			}
		}
		metricReplica = scaler.GlobalScaler.GetMetricReplica(modelReplicas, scaler.MaxStrategy)

		if history.CanTrain(now, conf.trainInterval) {
			err := p.Train(ctx)
			if err != nil {
				return metricReplica, err
			}
		}

	}
	return metricReplica, nil
}
