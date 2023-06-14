package scheduler

import (
	"context"
	"github.com/LL-res/AOM/common/errs"
	"github.com/LL-res/AOM/common/store"
	"github.com/LL-res/AOM/log"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/scaler"
	"github.com/LL-res/AOM/utils"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"sync"
	"time"
)

type Scheduler struct {
	Name     types.NamespacedName
	Interval time.Duration
}
type Conf struct {
	needTrain       bool
	trainInterval   time.Duration
	predictInterval time.Duration
}

var schedulers map[types.NamespacedName]*Scheduler

func GetOrNew(name types.NamespacedName, interval time.Duration) *Scheduler {
	if nil == schedulers {
		schedulers = make(map[types.NamespacedName]*Scheduler)
	}
	if nil == schedulers[name] {
		schedulers[name] = New(name, interval)
	}
	return schedulers[name]
}
func New(name types.NamespacedName, interval time.Duration) *Scheduler {
	return &Scheduler{
		Name: name,
		// the interval AOM call all the models
		Interval: interval,
	}
}

type ResPair struct {
	modelReplica []int32
	withModelKey string
}

func (s *Scheduler) DeepCopyInto(out *Scheduler) {

}
func (s *Scheduler) Run(ctx context.Context) {
	// interval 为instance配置
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	// TODO 获取当前时间判断是否需要对模型进行训练

	for _ = range ticker.C {
		waitGroup := sync.WaitGroup{}
		hide := store.GetHide(s.Name)
		hide.PredictorMap.Lock()
		scr := hide.Scaler
		ResChan := make(chan ResPair, len(hide.PredictorMap.Data))

		for withModelKey, pred := range hide.PredictorMap.Data {
			// 获取model以判断是否需要进行训练
			model, err := hide.ModelMap.Load(withModelKey)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			// 获取metric以判断该predictor所对应的metric
			metric, err := hide.MetricMap.Load(utils.GetNoModelKey(withModelKey))
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			// 进行预测
			waitGroup.Add(1)
			go func(withModelKey string, pred predictor.Predictor, scr *scaler.Scaler) {
				defer waitGroup.Done()
				pResult, err := pred.Predict(ctx)
				if err == errs.NO_SUFFICENT_DATA || err == errs.UNREADY_TO_PREDICT {
					log.Logger.Info("the predictor needs more metrics to be funtional", "predictor", withModelKey)
					return
				}
				if err != nil {
					log.Logger.Error(err, "predict failed", "predictor", withModelKey)
					return
				}
				targetVal, err := strconv.ParseFloat(metric.Target, 64)
				if err != nil {
					log.Logger.Error(err, "strconv failed")
					return
				}
				modelReplica, err := scr.GetModelReplica(pResult.PredictMetric, pResult.StartMetric, scaler.UnderThreshold, targetVal)
				if err != nil {
					log.Logger.Error(err, "get model replica failed", "key", withModelKey)
					return
				}
				ResChan <- ResPair{
					modelReplica: modelReplica,
					withModelKey: withModelKey,
				}
			}(withModelKey, pred, scr)
			if model.NeedTrain {
				// the err stands for if lastTime exists
				lastTime, err := hide.TrainHistory.Load(withModelKey)
				if !(err != nil || lastTime.Add(time.Second*time.Duration(model.UpdateInterval)).Before(time.Now())) {
					continue
				}
				// train use asynchronous,so it won`t block the process for too long
				if err := pred.Train(ctx); err != nil {
					log.Logger.Error(err, "train model failed", "key", withModelKey)
					continue
				}
			}
		}
		hide.PredictorMap.Unlock()
		waitGroup.Wait()

		//每一个metric对应的model的所有的预测副本数
		modelReplicas := make(map[string][][]int32)

		close(ResChan)
		// no prediction result yet
		if len(ResChan) == 0 {
			infos := make(map[string]int)
			for k, v := range hide.CollectorWorkerMap.Data {
				infos[k] = v.DataCap()
			}
			log.Logger.Info("no predictor results received", "metric cap", infos)
			continue
		}
		for pair := range ResChan {
			noModelKey := utils.GetNoModelKey(pair.withModelKey)
			if modelReplicas[noModelKey] == nil {
				modelReplicas[noModelKey] = [][]int32{pair.modelReplica}
			} else {
				modelReplicas[noModelKey] = append(modelReplicas[noModelKey], pair.modelReplica)
			}
		}
		log.Logger.Info("modelReplicas", "modelReplicas", modelReplicas)
		// 每一个metric对应的已经由model聚合完的副本数
		metricReplicas := make(map[string][]int32, 0)
		for noModelKey, modelReplica := range modelReplicas {
			metricReplicas[noModelKey] = scr.GetMetricReplica(modelReplica, scaler.MaxStrategy)
		}
		log.Logger.Info("metricReplicas", "metricReplicas", metricReplicas)
		//该数据结构对结果加权得出的结果进行暂存，以选出最后的扩所容副本数集合
		mReplicas := make([][]int32, 0, len(metricReplicas))
		for noModelKey, metricReplica := range metricReplicas {
			// 获取加权系数
			metric, err := hide.MetricMap.Load(noModelKey)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			utils.MulSlice(metric.Weight, metricReplica)
			mReplicas = append(mReplicas, metricReplica)
		}
		//扩所容副本选择集合
		objSet := utils.AddSlice(mReplicas...)
		//最终决定的扩容副本数，此刻的targetReplica并为除100，将除底数滞后以防止过多的类型转换
		targetReplica := scr.GetScaleReplica(objSet, scaler.SelectMax)
		log.Logger.Info("targetReplica", "targetReplica", targetReplica/100)
		if err := scr.UpTo(targetReplica / 100); err != nil {
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
