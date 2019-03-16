/*
Copyright 2018 Matt Moore

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

package v1alpha1

import (
	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WithPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WithPodSpec `json:"spec,omitempty"`
}

// PodSpeccable is implemented by types containing a PodTemplateSpec
// in the manner of ReplicaSet, Deployment, DaemonSet, StatefulSet.
type PodSpeccable corev1.PodTemplateSpec

type WithPodSpec struct {
	Template PodSpeccable `json:"template,omitempty"`
}

var _ apis.Validatable = (*WithPod)(nil)
var _ apis.Defaultable = (*WithPod)(nil)
var _ duck.Populatable = (*WithPod)(nil)
var _ duck.Implementable = (*PodSpeccable)(nil)

// Validate ensures WithPod is properly configured.
func (rt *WithPod) Validate() *apis.FieldError {
	return nil
}

var FreezeConfigMap = func(namespace, name string) string {
	return name
}

// SetDefaults ensures WithPod is properly configured.
func (rt *WithPod) SetDefaults() {
	for idx, v := range rt.Spec.Template.Spec.Volumes {
		// TODO(mattmoor): ProjectedVolumeSource
		if v.VolumeSource.ConfigMap == nil {
			continue
		}
		rt.Spec.Template.Spec.Volumes[idx].VolumeSource.ConfigMap.LocalObjectReference.Name =
			FreezeConfigMap(rt.Namespace, v.VolumeSource.ConfigMap.LocalObjectReference.Name)
	}
	for _, c := range rt.Spec.Template.Spec.InitContainers {
		for idx, env := range c.Env {
			if env.ValueFrom == nil || env.ValueFrom.ConfigMapKeyRef == nil {
				continue
			}
			c.Env[idx].ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name =
				FreezeConfigMap(rt.Namespace, env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name)
		}
	}
	for _, c := range rt.Spec.Template.Spec.Containers {
		for idx, env := range c.Env {
			if env.ValueFrom == nil || env.ValueFrom.ConfigMapKeyRef == nil {
				continue
			}
			c.Env[idx].ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name =
				FreezeConfigMap(rt.Namespace, env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name)
		}
	}
}

// GetFullType implements duck.Implementable
func (_ *PodSpeccable) GetFullType() duck.Populatable {
	return &WithPod{}
}

// Populate implements duck.Populatable
func (t *WithPod) Populate() {
	t.Spec.Template = PodSpeccable{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "container-name",
				Image: "container-image:latest",
			}},
		},
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WithPodList is a list of WithPod resources
type WithPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WithPod `json:"items"`
}
