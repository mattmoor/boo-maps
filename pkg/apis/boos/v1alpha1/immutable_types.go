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
	"github.com/knative/pkg/kmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImmutableMap is a specification for a ImmutableMap resource
type ImmutableMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec map[string]string `json:"spec"`
}

// Check that we can create OwnerReferences to a ImmutableMap.
var _ kmeta.OwnerRefable = (*ImmutableMap)(nil)
var _ apis.Validatable = (*ImmutableMap)(nil)
var _ apis.Defaultable = (*ImmutableMap)(nil)
var _ apis.Immutable = (*ImmutableMap)(nil)

func (r *ImmutableMap) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ImmutableMap")
}

// Validate ensures ImmutableMap is properly configured.
func (rt *ImmutableMap) Validate() *apis.FieldError {
	return nil
}

// CheckImmutableFields checks the immutable fields are not modified.
func (current *ImmutableMap) CheckImmutableFields(og apis.Immutable) *apis.FieldError {
	original, ok := og.(*ImmutableMap)
	if !ok {
		return &apis.FieldError{Message: "The provided original was not a ImmutableMap"}
	}

	if diff, err := kmp.SafeDiff(original.Spec, current.Spec); err != nil {
		return &apis.FieldError{
			Message: "Failed to diff ImmutableMap",
			Paths:   []string{"spec"},
			Details: err.Error(),
		}
	} else if diff != "" {
		return &apis.FieldError{
			Message: "Immutable fields changed (-old +new)",
			Paths:   []string{"spec"},
			Details: diff,
		}
	}
	return nil
}

// SetDefaults ensures ImmutableMap is properly configured.
func (rt *ImmutableMap) SetDefaults() {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImmutableMapList is a list of ImmutableMap resources
type ImmutableMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ImmutableMap `json:"items"`
}
