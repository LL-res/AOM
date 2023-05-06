/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/LL-res/AOM/collector"
	"github.com/LL-res/AOM/collector/prometheus_collector"
	"github.com/LL-res/AOM/common/basetype"
	"github.com/LL-res/AOM/common/consts"
	"github.com/LL-res/AOM/log"
	"github.com/LL-res/AOM/predictor"
	"github.com/LL-res/AOM/scheduler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	automationv1 "github.com/LL-res/AOM/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	defaultSyncPeriod       = 15 * time.Second
	defaultErrorRetryPeriod = 10 * time.Second
	metricMapKey            = "metricMap"
)

var (
	promcOnce  sync.Once
	pCollector collector.Collector
)

// AOMReconciler reconciles a AOM object
type AOMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=automation.buaa.io,resources=aoms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=automation.buaa.io,resources=aoms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=automation.buaa.io,resources=aoms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AOM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AOMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.Logger.WithName("reconcile")

	instance := &automationv1.AOM{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Info("instance deleted")
			return reconcile.Result{}, nil
		}
		logger.Error(err, "failed to get instance")
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}
	ctx = context.WithValue(ctx, consts.NAMESPACE, req.Namespace)

	handler := NewHandler(instance, r)

	if err := handler.Handle(ctx); err != nil {
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// 将这些predictor交给scheduler进行调度
	// 考虑不同的instance 对应不同的scheduler

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AOMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&automationv1.AOM{}).
		Complete(r)
}

func StartWorker(ctx context.Context, worker collector.MetricCollector, aom *automationv1.AOM, stopC chan struct{}) {
	ticker := time.NewTicker(aom.Spec.Collector.ScrapeInterval)
	end := false
	defer ticker.Stop()
	for _ = range ticker.C {
		if _, err := aom.CollectorWorkerMap.Load(worker.NoModelKey()); err != nil || end {
			break
		}
		select {
		case <-ctx.Done():
			log.Logger.Info("worker exit", "worker", worker.NoModelKey())
			end = true
			break
		case <-stopC:
			log.Logger.Info("worker exit", "worker", worker.NoModelKey())
			end = true
			break
		default:
			err := worker.Collect()
			if err != nil {
				log.Logger.Error(err, "fail to collect", "worker", worker.NoModelKey())
			}
		}
	}
}

type Handler struct {
	instance *automationv1.AOM
	//predictors map[*basetype.Metric][]predictor.Predictor
	*AOMReconciler
}

func NewHandler(instance *automationv1.AOM, reconciler *AOMReconciler) *Handler {
	return &Handler{
		instance:      instance,
		AOMReconciler: reconciler,
	}
}

func (hdlr *Handler) Handle(ctx context.Context) error {
	// 是由 status的更新导致
	if hdlr.instance.Status.Generation == hdlr.instance.Generation {
		return nil
	}
	// 防止过多层if嵌套
	var err error
	// create instance
	if hdlr.instance.Status.Generation == 0 {
		err = hdlr.handleCreate(ctx)
	}
	// update instance
	if hdlr.instance.Status.Generation != 0 &&
		hdlr.instance.Generation > hdlr.instance.Status.Generation {
		err = hdlr.handleUpdate(ctx)
	}
	if err != nil {
		return err
	}

	hdlr.instance.Status.Generation = hdlr.instance.Generation

	if err := hdlr.Status().Update(ctx, hdlr.instance); err != nil {
		log.Logger.Error(err, "update status failed")
		return err
	}
	return nil

}

func (hdlr *Handler) handleUpdate(ctx context.Context) error {
	if err := hdlr.handleCollector(ctx); err != nil {
		return err
	}
	if err := hdlr.handlePredictor(ctx); err != nil {
		return err
	}
	return nil
}

func (hdlr *Handler) handleCreate(ctx context.Context) error {
	promcOnce.Do(func() {
		pCollector = prometheus_collector.New()
	})
	err := pCollector.SetServerAddress(hdlr.instance.Spec.Collector.Address)
	if err != nil {
		log.Logger.Error(err, "fail to set collector server address")
		return err
	}
	// 初始化instance中的Map
	hdlr.instance.Init()
	if err := hdlr.handleCollector(ctx); err != nil {
		return err
	}
	//if hdlr.predictors == nil {
	//	hdlr.predictors = make(map[*basetype.Metric][]predictor.Predictor, 0)
	//}

	if err := hdlr.handlePredictor(ctx); err != nil {
		return err
	}

	if hdlr.instance.Scheduler == nil {
		hdlr.instance.Scheduler = scheduler.New(&hdlr.instance.Hide, hdlr.instance.Spec.Interval)
	}
	go hdlr.instance.Scheduler.Run(ctx)

	return nil
}

func (hdlr *Handler) handleDelete(ctx context.Context) error {
	return nil
}

func (hdlr *Handler) handleCollector(ctx context.Context) error {
	// 此操作为幂等操作
	// 其中的元素是格式化之后的metric，格式为: name/unit/query
	toDelete := make([]string, 0)
	toAdd := make([]basetype.Metric, 0)

	// spec中存在，但map中不存在，进行更新
	for metric := range hdlr.instance.Spec.Metrics {
		if _, ok := hdlr.instance.CollectorMap[metric.NoModelKey()]; !ok {
			toAdd = append(toAdd, metric)
		}
	}
	// map 中存在但 spec中不存在，进行删除
	for k := range hdlr.instance.CollectorMap {
		exist := false
		for metric := range hdlr.instance.Spec.Metrics {
			if metric.NoModelKey() == k {
				exist = true
				break
			}
		}
		if !exist {
			toDelete = append(toDelete, k)
		}
	}
	for _, v := range toDelete {
		// 对collecter worker进行退出控制
		close(hdlr.instance.CollectorMap[v])
		hdlr.instance.CollectorWorkerMap.Delete(v)
	}
	for _, m := range toAdd {
		pCollector.AddCustomMetrics(m)
		worker, err := pCollector.CreateWorker(m)
		if err != nil {
			log.Logger.Error(err, "fail to create metric collector worker")
			return err
		}
		hdlr.instance.CollectorWorkerMap.Store(m.NoModelKey(), worker)
		stopC := make(chan struct{})
		hdlr.instance.CollectorMap[m.NoModelKey()] = stopC
		go StartWorker(ctx, worker, hdlr.instance, stopC)
	}
	// 更新status
	hdlr.instance.Status.StatusCollectors = make([]automationv1.StatusCollector, 0, len(hdlr.instance.Spec.Metrics))
	for metric := range hdlr.instance.Spec.Metrics {
		// 此处仅作describe时显示
		hdlr.instance.Status.StatusCollectors = append(hdlr.instance.Status.StatusCollectors, automationv1.StatusCollector{
			Name:       metric.Name,
			Unit:       metric.Unit,
			Expression: metric.Query,
		})
	}
	if err := hdlr.Status().Update(ctx, hdlr.instance); err != nil {
		log.Logger.Error(err, "update status failed")
		return err
	}
	return nil
}

type mdlMtrc struct {
	basetype.Model
	basetype.Metric
}

func (hdlr *Handler) handlePredictor(ctx context.Context) error {
	// 扫一遍spec 查看现在所需的

	// sepc 中存在，map中不存在
	toAdd := make([]mdlMtrc, 0)
	for mtrc, mdls := range hdlr.instance.Spec.Metrics {
		for _, model := range mdls {
			if _, err := hdlr.instance.PredictorMap.Load(mtrc.WithModelKey(model.Type)); err != nil {
				toAdd = append(toAdd, mdlMtrc{
					Model:  model,
					Metric: mtrc,
				})
			}
		}
	}
	// map 中存在，spec中不存在
	toDelete := make([]string, 0)
	// 先将spec中的key都放入tempMap中，再进行比较以降低复杂度
	tempMap := make(map[string]struct{})
	for mtrc, mdls := range hdlr.instance.Spec.Metrics {
		for _, model := range mdls {
			tempMap[mtrc.WithModelKey(model.Type)] = struct{}{}
		}
	}
	for k := range hdlr.instance.PredictorMap.Data {
		if _, ok := tempMap[k]; !ok {
			toDelete = append(toDelete, k)
		}
	}
	for _, wmk := range toDelete {
		hdlr.instance.PredictorMap.Delete(wmk)
		//nmk := utils.GetNoModelKey(wmk)
		//找到metric对应的那一组predictor
		//metric, err := hdlr.instance.MetricMap.Get(nmk)
		//if err != nil {
		//	log.Logger.Error(err, "a must behaviour failed,predictor can not find the corresponding metric")
		//	return err
		//}
		//for i, pred := range hdlr.predictors[metric] {
		//	if pred.Key() == wmk {
		//		hdlr.predictors[metric] = append(hdlr.predictors[metric][:i], hdlr.predictors[metric][i+1:]...)
		//	}
		//}
	}
	for _, param := range toAdd {
		WithModelKey := param.WithModelKey(param.Model.Type)
		collect, err := hdlr.instance.CollectorWorkerMap.Load(param.NoModelKey())
		if err != nil {
			log.Logger.Error(err, "a must behaviour failed,predictor can not find the corresponding collector")
			return err
		}
		pred, err := predictor.NewPredictor(predictor.Param{
			WithModelKey:    WithModelKey,
			MetricCollector: collect,
			Model:           param.Model.Attr,
			ScaleTargetRef:  hdlr.instance.Spec.ScaleTargetRef,
		})
		if err != nil {
			log.Logger.Error(err, "new predictor failed")
			return err
		}
		hdlr.instance.PredictorMap.Store(param.WithModelKey(param.Type), pred)
		//metric, err := hdlr.instance.MetricMap.Get(param.NoModelKey())
		//if err != nil {
		//	log.Logger.Error(err, "a must behaviour failed,predictor can not find the corresponding metric")
		//	return err
		//}
		//hdlr.predictors[metric] = append(hdlr.predictors[metric], pred)

	}
	//TODO 展示部分的status还未完成
	if err := hdlr.Status().Update(ctx, hdlr.instance); err != nil {
		log.Logger.Error(err, "update status failed")
		return err
	}
	return nil
}
