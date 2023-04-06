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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	automationv1 "github.com/LL-res/AOM/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	defaultSyncPeriod       = 15 * time.Second
	defaultErrorRetryPeriod = 10 * time.Second
	metricMap               = "metricMap"
)

var (
	promcOnce  sync.Once
	pCollector collector.Collector
	logger     logr.Logger
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

	logger := log.FromContext(ctx)

	instance := &automationv1.AOM{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		logger.Error(err, "failed to get AOM")
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// TODO: validation

	scaleTargetRef := instance.Spec.ScaleTargetRef
	metricMap := make(map[string]struct{})
	ctx = context.WithValue(ctx, metricMap, metricMap)
	if err := checkCollector(ctx, instance); err != nil {
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AOMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&automationv1.AOM{}).
		Complete(r)
}
func checkCollector(ctx context.Context, aom *automationv1.AOM) error {
	logger := log.FromContext(ctx)

	if aom.Status.CollectorStatus == "up" {
		// 其中的元素是格式化之后的metric
		toDelete := make([]string, 0)
		toAdd := make([]automationv1.Metric, 0)

		// spec中存在，但map中不存在，进行更新
		for _, v := range aom.Spec.Metrics {
			if _, ok := aom.Status.CollectorMap[v.NoModelKey()]; !ok {
				toAdd = append(toAdd, v)
			}

		}
		// map 中存在但 spec中不存在，进行删除
		for k := range aom.Status.CollectorMap {
			exist := false
			for _, v := range aom.Spec.Metrics {
				if v.NoModelKey() == k {
					exist = true
					break
				}
			}
			if !exist {
				toDelete = append(toDelete)
			}
			toDelete = append(toDelete, k)
		}
		metricMap := ctx.Value(metricMap).(map[string]struct{})
		for _, v := range toDelete {
			delete(metricMap, v)
		}
		for _, v := range toAdd {
			metricType := collector.MetricType{
				Name: v.Name,
				Unit: v.Unit,
			}
			pCollector.AddCustomMetrics(metricType, v.Query)
			worker, err := pCollector.CreateWorker(metricType)
			if err != nil {
				return err
			}
			go StartWorker(ctx, worker, aom)
		}
		// 更新status
		aom.Status.CollectorMap = make(map[string]struct{})
		for _, v := range aom.Spec.Metrics {
			aom.Status.CollectorMap[v.NoModelKey()] = struct{}{}
		}
		return nil
	}

	promcOnce.Do(func() {
		pCollector = prometheus_collector.New()
	})
	err := pCollector.SetServerAddress(aom.Spec.Collector.Address)
	if err != nil {
		logger.Error(err, "fail to set collector server address")
		return err
	}

	for _, metric := range aom.Spec.Metrics {
		ctx.Value(metricMap).(map[string]struct{})[metric.NoModelKey()] = struct{}{}
		pCollector.AddCustomMetrics(collector.MetricType{
			Name: metric.Name,
			Unit: metric.Unit,
		}, metric.Query)
	}
	targets := pCollector.ListMetricTypes()
	workers := make([]collector.MetricCollector, 0)
	for _, target := range targets {
		worker, err := pCollector.CreateWorker(target)
		if err != nil {
			return err
		}
		workers = append(workers, worker)
	}
	for _, worker := range workers {
		go StartWorker(ctx, worker, aom)
	}

	aom.Status.CollectorStatus = "up"

	return nil
}
func StartWorker(ctx context.Context, worker collector.MetricCollector, aom *automationv1.AOM) {
	ticker := time.NewTicker(aom.Spec.Collector.ScrapeInterval)
	end := false
	defer ticker.Stop()
	for _ = range ticker.C {
		if _, ok := ctx.Value(metricMap).(map[string]struct{})[worker.String()]; !ok || end {
			break
		}
		select {
		case <-ctx.Done():
			logger.Info("worker exit", "worker", worker)
			end = true
			break
		default:
			err := worker.Collect()
			if err != nil {
				logger.Error(err, "fail to collect",
					"worker", worker)
			}
		}
	}
}
