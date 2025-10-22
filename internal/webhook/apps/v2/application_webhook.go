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

package v2

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv2 "github.com/wuyong7240/application-operator-plus/api/apps/v2"
)

// nolint:unused
// log is for logging in this package.
var applicationlog = logf.Log.WithName("application-resource")

// SetupApplicationWebhookWithManager registers the webhook for Application in the manager.
func SetupApplicationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&appsv2.Application{}).
		WithValidator(&ApplicationCustomValidator{
			DefaultDeploymentReplicasMax: 10,
		}).
		WithDefaulter(&ApplicationCustomDefaulter{
			DefaultDeploymentReplicas: 3,
		}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// +kubebuilder:webhook:path=/mutate-apps-wuyong-cn-v2-application,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.wuyong.cn,resources=applications,verbs=create;update,versions=v2,name=mapplication-v2.kb.io,admissionReviewVersions=v1,matchPolicy=Exact

type ApplicationCustomDefaulter struct {
	DefaultDeploymentReplicas int32
}

var _ webhook.CustomDefaulter = &ApplicationCustomDefaulter{}

func (d *ApplicationCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	application, ok := obj.(*appsv2.Application)
	if !ok {
		return fmt.Errorf("expected an application object but got %T", obj)
	}
	applicationlog.Info("Defaulting for Application", "name", application.Name)

	if application.Spec.Workflow.Replicas == nil {
		application.Spec.Workflow.Replicas = new(int32)
		*application.Spec.Workflow.Replicas = d.DefaultDeploymentReplicas
	}

	return nil
}

// validation
// +kubebuilder:webhook:path=/validate-apps-wuyong-cn-v2-application,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.wuyong.cn,resources=applications,verbs=create;update,versions=v2,name=vapplication-v2.kb.io,admissionReviewVersions=v1,matchPolicy=Exact

type ApplicationCustomValidator struct {
	DefaultDeploymentReplicasMax int32
}

var _ webhook.CustomValidator = &ApplicationCustomValidator{}

func (v *ApplicationCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*appsv2.Application)
	if !ok {
		return nil, fmt.Errorf("expected an Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon Creation", "name", application.Name)

	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v2 Errors!"}, err
	}
	return nil, nil
}

func (v *ApplicationCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	application, ok := newObj.(*appsv2.Application)
	if !ok {
		return nil, fmt.Errorf("expected an Application object but got %T", newObj)
	}
	applicationlog.Info("Validation for Application upon Update", "name", application.Name)

	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v2 Errors!"}, err
	}
	return nil, nil
}

func (v *ApplicationCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*appsv2.Application)
	if !ok {
		return nil, fmt.Errorf("expected an Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon Update", "name", application.Name)

	if err := v.validateApplication(application); err != nil {
		return admission.Warnings{"Application Webhook v2 Errors!"}, err
	}
	return nil, nil
}

func (v *ApplicationCustomValidator) validateApplication(application *appsv2.Application) error {
	if *application.Spec.Workflow.Replicas > v.DefaultDeploymentReplicasMax {
		return fmt.Errorf("replicas too many error")
	}
	return nil
}
