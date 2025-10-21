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

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1 "github.com/wuyong7240/application-operator-plus/api/apps/v1"
)

// nolint:unused
// log is for logging in this package.
var applicationlog = logf.Log.WithName("application-resource")

// SetupApplicationWebhookWithManager registers the webhook for Application in the manager.
func SetupApplicationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&appsv1.Application{}).
		WithValidator(&ApplicationCustomValidator{
			DefaultDeploymentReplicasMax: 10,
		}).
		WithDefaulter(&ApplicationCustomDefaulter{
			DefaultDeploymentReplicas: 3,
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-apps-wuyong-cn-v1-application,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.wuyong.cn,resources=applications,verbs=create;update,versions=v1,name=mapplication-v1.kb.io,admissionReviewVersions=v1

// ApplicationCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Application when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ApplicationCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
	DefaultDeploymentReplicas int32
}

var _ webhook.CustomDefaulter = &ApplicationCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Application.
func (d *ApplicationCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	// 从上下文中，获取对应的Application对象
	application, ok := obj.(*appsv1.Application)

	if !ok {
		return fmt.Errorf("expected an Application object but got %T", obj)
	}
	applicationlog.Info("Defaulting for Application", "name", application.GetName())

	// TODO(user): fill in your defaulting logic.
	// 查看Application对象的Replicas是否为空，如果为空，设置为3
	if application.Spec.Deployment.Replicas == nil {
		application.Spec.Deployment.Replicas = new(int32)
		*application.Spec.Deployment.Replicas = d.DefaultDeploymentReplicas
	}

	return nil
}

// func (r *Application) Default() {
// 	applicationlog.Info("default", "name", r.Name)

// 	if r.Spec.Deployment.Replicas == nil {
// 		r.Spec.Deployment.Replicas = new(int32)
// 		*r.Spec.Deployment.Replicas = 2
// 	}
// }

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-apps-wuyong-cn-v1-application,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.wuyong.cn,resources=applications,verbs=create;update,versions=v1,name=vapplication-v1.kb.io,admissionReviewVersions=v1

// ApplicationCustomValidator struct is responsible for validating the Application resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ApplicationCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
	DefaultDeploymentReplicasMax int32
}

var _ webhook.CustomValidator = &ApplicationCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*appsv1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon creation", "name", application.GetName())

	// TODO(user): fill in your validation logic upon object creation.
	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v1 Errors!"}, err
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	application, ok := newObj.(*appsv1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object for the newObj but got %T", newObj)
	}
	applicationlog.Info("Validation for Application upon update", "name", application.GetName())

	// TODO(user): fill in your validation logic upon object update.
	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v1 Errors!"}, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*appsv1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon deletion", "name", application.GetName())

	// TODO(user): fill in your validation logic upon object deletion.
	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v1 Errors!"}, err
	}

	return nil, nil
}

func (v *ApplicationCustomValidator) validateApplication(application *appsv1.Application) error {
	if *application.Spec.Deployment.Replicas > v.DefaultDeploymentReplicasMax {
		return fmt.Errorf("replicas too many error")
	}
	return nil
}
