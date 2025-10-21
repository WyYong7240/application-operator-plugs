package controller

import (
	"context"
	"reflect"

	v2 "github.com/wuyong7240/application-operator-plus/api/apps/v2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, app *v2.Application) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 先根据Application中的Namespace和Name信息查询对应的Deployment是否存在
	var dp = &appsv1.Deployment{}
	// types.NamespaceedName用于唯一标识Kubernetes集群中的资源,dp是一个指针，如果Get方法成功执行，这个指针指向从API服务器中获取的Deployment对象
	err := r.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Name,
	}, dp)

	// 没有错误发生时，更新状态
	if err == nil {
		log.Info("The Deployment has already exist.")
		// 使用reflect.DeepEqual比较Deployment的状态(dp.Status)与Application自定义资源的工作流状态(app.Status.Workflow)是否完全相同
		// 如果相同，说明没有变化，不需要进一步处理
		if reflect.DeepEqual(dp.Status, app.Status.Workflow) {
			return ctrl.Result{}, nil
		}

		// 如果不同，需要更新Application的状态
		app.Status.Workflow = dp.Status
		// 调用r.Status().Update更新Application资源的状态
		if err := r.Status().Update(ctx, app); err != nil {
			log.Error(err, "Failed to update Application status")
			// 返回一个带有重新排队时间的结果和错误，表示需要在一段时间后重试
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		log.Info("The Application status has been updated.")
		return ctrl.Result{}, nil
	}

	// 如果不是NotFound的错误，即发生了其他错误，结束本轮调谐，一段时间后重试
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Deployment, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 根据Application资源实例信息来构造Deployment实例
	newDp := &appsv1.Deployment{}
	newDp.SetName(app.Name)
	newDp.SetNamespace(app.Namespace)
	newDp.SetLabels(app.Labels)
	newDp.Spec = app.Spec.Workflow.DeploymentSpec
	// 这是Pod的模板，Pod模板的Labels是独立的，必须单独设置，如果不设置，回到是Deployment的selector无法匹配到Pod
	newDp.Spec.Template.SetLabels(app.Labels)

	// 用于建立App里擦同与Deployment之间的父子关系：Kubernetes通过owner Reference实现级联删除，当Application被删除时，Kubernetes
	// 会自动删除它创建的Deployment; r.scheme用来识别资源类型的Scheme，确保类型正确
	if err := ctrl.SetControllerReference(app, newDp, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// 在集群中创建Deployment：调用客户端的Create方法，将newDp提交到Kubernetes API Server
	if err := r.Create(ctx, newDp); err != nil {
		log.Error(err, "Failed to create Deployment, will requeue, after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	log.Info("The Deployment has been created.")
	return ctrl.Result{}, nil
}
