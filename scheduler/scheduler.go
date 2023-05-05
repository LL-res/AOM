package scheduler

import (
	"context"
	aomtype "github.com/LL-res/AOM/common/aomtype"
	"github.com/LL-res/AOM/log"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/scaler"
	"github.com/LL-res/AOM/utils"
	"sync"
	"time"
)

var (
	GlobalScheduler *Scheduler
)

type Scheduler struct {
	*aomtype.Hide
	interval time.Duration
}
type Conf struct {
	needTrain       bool
	trainInterval   time.Duration
	predictInterval time.Duration
}

func New(Hide *aomtype.Hide, interval time.Duration) *Scheduler {
	return &Scheduler{
		Hide:     Hide,
		interval: interval,
	}
}

type ResPair struct {
	modelReplica []int32
	withModelKey string
}

func (s *Scheduler) Run(ctx context.Context) {
	// interval 为instance配置
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	// TODO 获取当前时间判断是否需要对模型进行训练
	for _ = range ticker.C {
		waitGroup := sync.WaitGroup{}

		s.Hide.PredictorMap.Lock()

		ResChan := make(chan ResPair, len(s.Hide.PredictorMap.Data))

		for withModelKey, pred := range s.Hide.PredictorMap.Data {
			// 获取model以判断是否需要进行训练
			model, err := s.Hide.ModelMap.Load(withModelKey)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			// 获取metric以判断该predictor所对应的metric
			metric, err := s.Hide.MetricMap.Load(utils.GetNoModelKey(withModelKey))
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			go func(withModelKey string, pred predictor.Predictor) {
				waitGroup.Add(1)
				defer waitGroup.Done()
				pResult, err := pred.Predict(ctx)
				if err != nil {
					log.Logger.Error(err, "predict failed", "predictor", withModelKey)
					return
				}
				modelReplica, err := scaler.GlobalScaler.GetModelReplica(pResult.PredictMetric, pResult.StartMetric, scaler.Steady, metric.Target)
				if err != nil {
					log.Logger.Error(err, "get model replica failed", "key", withModelKey)
					return
				}
				ResChan <- ResPair{
					modelReplica: modelReplica,
					withModelKey: withModelKey,
				}
			}(withModelKey, pred)
			if model.NeedTrain {
				if err := pred.Train(ctx); err != nil {
					log.Logger.Error(err, "train model failed", "key", withModelKey)
					continue
				}
			}
		}
		s.Hide.PredictorMap.Unlock()
		waitGroup.Wait()

		//每一个metric对应的model的所有的预测副本数
		modelReplicas := make(map[string][][]int32)

		close(ResChan)
		for pair := range ResChan {
			noMoedlKey := utils.GetNoModelKey(pair.withModelKey)
			if modelReplicas[noMoedlKey] == nil {
				modelReplicas[noMoedlKey] = [][]int32{pair.modelReplica}
			} else {
				modelReplicas[noMoedlKey] = append(modelReplicas[noMoedlKey], pair.modelReplica)
			}
		}
		// 每一个metric对应的已经由model聚合完的副本数
		metricReplicas := make(map[string][]int32, 0)
		for noModelKey, modelReplica := range modelReplicas {
			metricReplicas[noModelKey] = scaler.GlobalScaler.GetMetricReplica(modelReplica, scaler.MaxStrategy)
		}
		//该数据结构对结果加权得出的结果进行暂存，以选出最后的扩所容副本数集合
		mReplicas := make([][]int32, 0, len(metricReplicas))
		for noModelKey, metricReplica := range metricReplicas {
			// 获取加权系数
			metric, err := s.Hide.MetricMap.Load(noModelKey)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			utils.MulSlice(metric.Weight, metricReplica)
			mReplicas = append(mReplicas, metricReplica)
		}
		//扩所容副本选择集合
		objSet := utils.AddSlice(mReplicas...)
		//最终决定的扩容副本数
		targetReplica := scaler.GlobalScaler.GetScaleReplica(objSet, scaler.SelectMax)
		if err := scaler.GlobalScaler.UpTo(targetReplica); err != nil {
			log.Logger.Error(err, "scale up failed")
		}
	}
}

//// predictors : 每个metric指标对应一组predictor，predictors中包含一个aom实例所拥有的全部的predictor，并按所属metric不同，分为不同的组
//func (s *Scheduler) HandlePredictors(ctx context.Context, predictors map[automationv1.Metric][]predictor.Predictor) error {
//	// 何时预测，何时更新的信息记录在model中
//	// 生成一个map，记录predictor与model的映射关系
//	// status 里应该记录历史数据，与时间戳，以检查是否需要进行操作
//	metricReplicas := make([][]int32, 0)
//	for m, p := range predictors {
//		metricReplica, err := s.HandleForOneMetric(ctx, m, p)
//		if err != nil {
//			log.Println(err)
//			continue
//		}
//		metricReplicas = append(metricReplicas, metricReplica)
//	}
//	objSet := scaler.GlobalScaler.GetObjReplica(metricReplicas, scaler.MaxStrategy)
//	targetReplica := scaler.GlobalScaler.GetScaleReplica(objSet, scaler.SelectMax)
//
//	if err := scaler.GlobalScaler.UpTo(targetReplica); err != nil {
//		log.Println(err)
//		return err
//	}
//	return nil
//}
//
//func (s *Scheduler) HandleForOneMetric(ctx context.Context, metric automationv1.Metric, predictors []predictor.Predictor) ([]int32, error) {
//
//	metricReplica := make([]int32, 0)
//
//	for _, p := range predictors {
//		history, ok := s.aom.Status.PredictorHistory.Load(p.Key())
//		if !ok {
//			// 此刻status中还没有任何history存在说明此刻predictor还未进行过动作,则应进行初始化
//			s.aom.Status.PredictorHistory.Store(p.Key(), new(automationv1.PredictorHistory))
//		}
//		model, ok := predictor.PredictorModelMap.Load(p.Key())
//		if !ok {
//			return nil, errors.New("model not find")
//		}
//		//拿到了model与history，history中包含了历史记录
//		//model中包含着对predictor进行调用的时间戳
//		conf := Conf{
//			needTrain:       false,
//			predictInterval: model.PredcitInterval.Duration,
//		}
//		switch model.Type {
//		case automationv1.TypeGRU, automationv1.TypeLSTM:
//			conf.trainInterval = model.GRU.UpdateInterval.Duration
//			conf.needTrain = true
//		}
//		now := time.Now()
//
//		modelReplicas := make([][]int32, 0)
//		if history.CanPredict(now, conf.predictInterval) {
//			var err error
//			pResult, err := p.Predict(ctx, s.aom)
//			if err != nil { // 此时的错误包含了数据不足，模型未训练等
//				log.Println(err)
//				continue
//				//此时仅对err进行打印
//			} else {
//				temp, err := scaler.GlobalScaler.GetModelReplica(pResult.StartMetric, scaler.Steady, metric.Target)
//				if err != nil {
//					log.Println(err)
//					continue
//				}
//				modelReplicas = append(modelReplicas, temp)
//			}
//		}
//		metricReplica = scaler.GlobalScaler.GetMetricReplica(modelReplicas, scaler.MaxStrategy)
//
//		if history.CanTrain(now, conf.trainInterval) {
//			err := p.Train(ctx)
//			if err != nil {
//				return metricReplica, err
//			}
//		}
//
//	}
//	return metricReplica, nil
//}
