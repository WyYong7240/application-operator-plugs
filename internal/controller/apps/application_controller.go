/*
Copyright 2025 wuyong.

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

package controller

import (
	"context"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "github.com/wuyong7240/application-operator-plus/api/apps/v1"
	v2 "github.com/wuyong7240/application-operator-plus/api/apps/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.wuyong.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.wuyong.cn,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.wuyong.cn,resources=applications/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile

var CounterReconcileApplication int64

// 表示通用的重新排队时间间隔
const GenericRequeueDuration = 1 * time.Minute

// req表示需要调谐的资源，包含Namespace和Name, ctrl.Result控制是否重试，延迟重试多久，如果error非nil会触发重试
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	<-time.NewTicker(100 * time.Millisecond).C
	// 获取日志记录器，便于在不同reconcile调用中区分日志来源
	log := log.FromContext(ctx)

	// 用于统计Reconcile被调用了多少次
	CounterReconcileApplication += 1
	log.Info("Starting a reconcile", "number", CounterReconcileApplication)

	app := &v2.Application{}
	// 从API Server中获取Application实例
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application not found.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get the Application, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// reconcile sub-resource, 调谐子资源
	var result ctrl.Result
	var err error

	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Deployment.")
		return result, err
	}

	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcle Service.")
		return result, err
	}

	log.Info("All resources have been reconciled.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("Setup")

	return ctrl.NewControllerManagedBy(mgr).
		// 监听Application资源，通过predicate.Funcs自定义哪些事件会触发Reconcile
		For(&v1.Application{}, builder.WithPredicates(predicate.Funcs{
			// 一旦创建Application，立即触发Reconcile
			CreateFunc: func(event event.CreateEvent) bool {
				return true
			},
			// Application被删除，但是不触发Reconcile，仅打印日志，不会执行任何清理逻辑
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Application has been deleted.", "Name", event.Object.GetName())
				return false
			},
			// 只有当ResourceVersion不同，且Spec发生变化时，才触发Reconcile
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(event.ObjectNew.(*v1.Application).Spec, event.ObjectOld.(*v1.Application).Spec) {
					return false
				}
				return true
			},
		})).
		// 监听Deployment子资源
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			// 由于是控制器自己创建的，无需响应
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			// 当Deployment被删除（例如误删）,触发Reconcile，让控制器重新创建Deployment，实现自愈
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Deployment has been deleted.", "Name", event.Object.GetName())
				return true
			},
			// 只有Spec变化时才触发，防止状态同步风暴，如果Deployment.Spec被外部修改，控制器会将其纠正回期望状态
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(event.ObjectNew.(*appsv1.Deployment).Spec, event.ObjectOld.(*appsv1.Deployment).Spec) {
					return false
				}
				return true
			},
			GenericFunc: nil,
		})).
		// 监听Service资源,与Deployment资源类似
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Service has been deleted.", "Name", event.Object.GetName())
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(event.ObjectNew.(*v1.Application).Spec, event.ObjectOld.(*v1.Application).Spec) {
					return false
				}
				return true
			},
		})).
		// 给控制器起名，日志和metrics中显示为controller "application"
		Named("application").
		// 完成注册，将Reconciler绑定到控制器上，并启动事件监听
		Complete(r)
}
