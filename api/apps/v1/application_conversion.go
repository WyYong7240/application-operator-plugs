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
	"log"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	appsv2 "github.com/wuyong7240/application-operator-plus/api/apps/v2"
)

// ConvertTo converts this Application (v1) to the Hub version (v2).
func (src *Application) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*appsv2.Application)
	log.Printf("ConvertTo: Converting Application from Spoke version v1 to Hub version v2;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

	// TODO(user): Implement conversion logic from v1 to v2
	// Example: Copying Spec fields
	// dst.Spec.Size = src.Spec.Replicas

	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Service = src.Spec.Service
	dst.Spec.Workflow = src.Spec.Deployment

	// Status
	dst.Status.Network = src.Status.Network
	dst.Status.Workflow = src.Status.Workflow

	return nil
}

// ConvertFrom converts the Hub version (v2) to this Application (v1).
func (dst *Application) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*appsv2.Application)
	log.Printf("ConvertFrom: Converting Application from Hub version v2 to Spoke version v1;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

	// TODO(user): Implement conversion logic from v2 to v1
	// Example: Copying Spec fields
	// dst.Spec.Replicas = src.Spec.Size

	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Deployment = src.Spec.Workflow
	dst.Spec.Service = src.Spec.Service

	// Status
	dst.Status.Network = src.Status.Network
	dst.Status.Workflow = src.Status.Workflow

	return nil
}
