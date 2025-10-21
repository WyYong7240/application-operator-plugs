package controller

import (
	"context"
	"reflect"

	v2 "github.com/wuyong7240/application-operator-plus/api/apps/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ApplicationReconciler) reconcileService(ctx context.Context, app *v2.Application) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	// 根据Application的Namespace和Name信息来查询对应的Servce资源
	var svc = &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, svc)

	// 如果查到了对应的Service
	if err == nil {
		log.Info("The Service has already exist.")
		// 利用reflect.DeepEqual判断现存的Service状态相比之前的Application中的状态是否有更新，如果有就更新Application中的Service状态
		if reflect.DeepEqual(svc.Status, app.Status.Network) {
			return ctrl.Result{}, err
		}

		// Service状态发生变化，将现存的Service状态赋值给Application中的Service状态，更新Application对Service的追踪
		app.Status.Network = svc.Status
		// 调用更新
		if err = r.Status().Update(ctx, app); err != nil {
			log.Error(err, "Failed to update Application Status")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("The Application status has been updated.")
		return ctrl.Result{}, nil
	}

	// 如果不是Not Found的错误，间隔一段时间后重试
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Service, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 如果是Not Found的错误，根据Application中的ServiceSpec，创建一个新的Service
	newSvc := &corev1.Service{}
	newSvc.SetName(app.Name)
	newSvc.SetNamespace(app.Namespace)
	newSvc.SetLabels(app.Labels)
	newSvc.Spec = app.Spec.Service.ServiceSpec
	newSvc.Spec.Selector = app.Labels

	// 设置所有者引用，将Application设置为Service的所有者，
	// 当Application被删除时，Service会被自动删除
	if err := ctrl.SetControllerReference(app, newSvc, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 调用客户端API创建Service资源
	if err := r.Create(ctx, newSvc); err != nil {
		log.Error(err, "Faield to create Service, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	log.Info("The Service has been created.")
	return ctrl.Result{}, nil
}
