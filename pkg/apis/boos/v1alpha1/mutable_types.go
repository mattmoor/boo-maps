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
	"github.com/knative/pkg/kmeta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MutableMap is a specification for a MutableMap resource
type MutableMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec map[string]string `json:"spec"`
}

// Check that we can create OwnerReferences to a MutableMap.
var _ kmeta.OwnerRefable = (*MutableMap)(nil)
var _ apis.Validatable = (*MutableMap)(nil)
var _ apis.Defaultable = (*MutableMap)(nil)

func (r *MutableMap) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("MutableMap")
}

// Validate ensures MutableMap is properly configured.
func (rt *MutableMap) Validate() *apis.FieldError {
	return nil
}

// SetDefaults ensures MutableMap is properly configured.
func (rt *MutableMap) SetDefaults() {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MutableMapList is a list of MutableMap resources
type MutableMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MutableMap `json:"items"`
}
