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

package immutable

import (
	"context"
	"fmt"

	"github.com/knative/pkg/controller"
	"github.com/knative/serving/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
	clientset "github.com/mattmoor/boo-maps/pkg/client/clientset/versioned"
	boosscheme "github.com/mattmoor/boo-maps/pkg/client/clientset/versioned/scheme"
	informers "github.com/mattmoor/boo-maps/pkg/client/informers/externalversions/boos/v1alpha1"
	listers "github.com/mattmoor/boo-maps/pkg/client/listers/boos/v1alpha1"
	"github.com/mattmoor/boo-maps/pkg/reconciler/immutable/resources"
	"github.com/mattmoor/boo-maps/pkg/reconciler/immutable/resources/names"
)

const controllerAgentName = "immutable-controller"

// Reconciler is the controller implementation for Filter resources
type Reconciler struct {
	*reconciler.Base

	boosclientset clientset.Interface

	immutableMapLister listers.ImmutableMapLister
	configMapLister    corev1listers.ConfigMapLister
}

// Check that we implement the controller.Reconciler interface.
var _ controller.Reconciler = (*Reconciler)(nil)

func init() {
	// Add immutable-controller types to the default Kubernetes Scheme so Events can be
	// logged for immutable-controller types.
	boosscheme.AddToScheme(scheme.Scheme)
}

// NewController returns a new immutable controller
func NewController(
	opt reconciler.Options,
	boosclientset clientset.Interface,
	immutableMapInformer informers.ImmutableMapInformer,
	configMapInformer corev1informers.ConfigMapInformer,
) *controller.Impl {
	r := &Reconciler{
		Base:               reconciler.NewBase(opt, controllerAgentName),
		boosclientset:      boosclientset,
		immutableMapLister: immutableMapInformer.Lister(),
		configMapLister:    configMapInformer.Lister(),
	}
	impl := controller.NewImpl(r, r.Logger, "ImmutableMaps",
		reconciler.MustNewStatsReporter("ImmutableMaps", r.Logger))

	r.Logger.Info("Setting up event handlers")

	// Set up an event handler for when ImmutableMap resources change.
	immutableMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    impl.Enqueue,
		UpdateFunc: controller.PassNew(impl.Enqueue),
		DeleteFunc: impl.Enqueue,
	})

	// Set up an event handler for when Knative Service resources that we own change.
	configMapInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ImmutableMap")),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    impl.EnqueueControllerOf,
			UpdateFunc: controller.PassNew(impl.EnqueueControllerOf),
			DeleteFunc: impl.EnqueueControllerOf,
		},
	})

	return impl
}

// Reconcile implements controller.Reconciler
func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the filter resource with this namespace/name
	original, err := c.immutableMapLister.ImmutableMaps(namespace).Get(name)
	if errors.IsNotFound(err) {
		// The ImmutableMap resource may no longer exist, in which case we stop processing.
		runtime.HandleError(fmt.Errorf("filter %q in work queue no longer exists", key))
		return nil
	} else if err != nil {
		return err
	}
	im := original.DeepCopy()

	// Reconcile this copy of the filter and then write back any status
	// updates regardless of whether the reconciliation errored out.
	return c.reconcile(ctx, im)
}

func (c *Reconciler) reconcile(ctx context.Context, im *v1alpha1.ImmutableMap) error {
	if err := c.reconcileConfigMap(ctx, im); err != nil {
		return err
	}
	return nil
}

func (c *Reconciler) reconcileConfigMap(ctx context.Context, im *v1alpha1.ImmutableMap) error {
	cmName := names.ConfigMap(im)
	cm, err := c.configMapLister.ConfigMaps(im.Namespace).Get(cmName)
	if apierrs.IsNotFound(err) {
		desiredCM := resources.MakeConfigMap(im)
		cm, err = c.KubeClientSet.CoreV1().ConfigMaps(im.Namespace).Create(desiredCM)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		desiredCM := resources.MakeConfigMap(im)
		if !equality.Semantic.DeepEqual(cm.Data, desiredCM.Data) {
			cm = cm.DeepCopy()
			cm.Data = desiredCM.Data
			cm, err = c.KubeClientSet.CoreV1().ConfigMaps(im.Namespace).Update(cm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
