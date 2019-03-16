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

package resources

import (
	"github.com/knative/pkg/kmeta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
	"github.com/mattmoor/boo-maps/pkg/reconciler/mutable/resources/names"
)

func MakeImmutableMap(im *v1alpha1.MutableMap) *v1alpha1.ImmutableMap {
	return &v1alpha1.ImmutableMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.ImmutableMap(im),
			Namespace:       im.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(im)},
			Annotations:     im.ObjectMeta.Annotations,
		},
		Spec: im.Spec,
	}
}
