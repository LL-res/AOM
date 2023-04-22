package scheduler

import (
	"context"
	"errors"
	automationv1 "github.com/LL-res/AOM/api/v1"
	"github.com/LL-res/AOM/predictor"
	"log"
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
		history, ok := s.aom.Status.PredictorHistory.Load(p.Key())
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

		replicas := make([]int32, 0)
		totalReplicas := make([][]int32, 0)
		if history.CanPredict(now, conf.predictInterval) {
			var err error
			replicas, err = p.Predict(ctx, s.aom)
			if err != nil { // 此时的错误包含了数据不足，模型未训练等
				log.Println(err)
				//此时仅对err进行打印
			} else {
				statusHistory, _ := s.aom.Status.PredictorHistory.Load(p.Key())
				// 添加新的历史数据应放在刚刚结束动作的代码块处
				statusHistory.AppendPredictorHistory(now)
				temp := make([]int32, len(replicas))
				copy(temp, replicas)
				totalReplicas = append(totalReplicas, temp)
			}
		}

		// predictor可考虑只返回预测指标，把计算副本数等部分拆出交给scaler来做

		if history.CanTrain(now, conf.trainInterval) {
			err := p.Train(ctx)
			if err != nil {
				return err
			}
		}

	}

}
