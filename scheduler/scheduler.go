package scheduler

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/predictor"
	"time"
)

type Scheduler struct {
	aom *automationv1.AOM
}
type Conf struct {
	needTrain       bool
	trainInterval   time.Duration
	predictInterval time.Duration
}

func (s *Scheduler) HandlePredictors(ctx context.Context, predictors []predictor.Predictor) error {
	// 何时预测，何时更新的信息记录在model中
	// 生成一个map，记录predictor与model的映射关系
	// status 里应该记录历史数据，与时间戳，以检查是否需要进行操作
	for _, p := range predictors {
		history, ok := s.aom.Status.PredictorHistory[p.Key()]
		if !ok {
			return errors.New("history not find")
		}
		model, ok := predictor.PredictorModelMap.Load(p.Key())
		if !ok {
			return errors.New("model not find")
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
		if history.CanPredict(now, conf.predictInterval) {
			//此处考虑进行下沉
			replicas, err := p.Predict(ctx, s.aom)
			if err != nil {
				return err
			}
		}
		s.aom.Status.PredictorHistory[p.Key()].AppendPredictorHistory(now)
		// 还需判断是否出于资源利用低谷
		if history.CanTrain(now, conf.trainInterval) {
			err := p.Train(ctx)
			if err != nil {
				return err
			}
		}

	}
}
