// +k8s:deepcopy-gen=package
package shared

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type DeploymentTemplate struct {
	appsv1.DeploymentSpec `json:",inline"`
}

type ServiceTemplate struct {
	corev1.ServiceSpec `json:",inline"`
}
