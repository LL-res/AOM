package scheduler

import (
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/predictor"
)

type Scheduler struct {
	aom *automationv1.AOM
}

func (s *Scheduler) HandlePredictors(predictors []predictor.Predictor) error {
	// 何时预测，何时更新的信息记录在model中
	// 生成一个map，记录predictor与model的映射关系
	// status 里应该记录历史数据，与时间戳，以检查是否需要进行操作
	predictor.PredictorModelMap
	for _, p := range predictors {
		s.aom.Status
	}
}
