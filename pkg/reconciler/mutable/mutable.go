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

package mutable

import (
	"context"
	"fmt"

	"github.com/knative/pkg/controller"
	"github.com/knative/serving/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"

	"github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
	clientset "github.com/mattmoor/boo-maps/pkg/client/clientset/versioned"
	boosscheme "github.com/mattmoor/boo-maps/pkg/client/clientset/versioned/scheme"
	informers "github.com/mattmoor/boo-maps/pkg/client/informers/externalversions/boos/v1alpha1"
	listers "github.com/mattmoor/boo-maps/pkg/client/listers/boos/v1alpha1"
	"github.com/mattmoor/boo-maps/pkg/reconciler/mutable/resources"
	"github.com/mattmoor/boo-maps/pkg/reconciler/mutable/resources/names"
)

const controllerAgentName = "mutable-controller"

// Reconciler is the controller implementation for Filter resources
type Reconciler struct {
	*reconciler.Base

	boosclientset clientset.Interface

	mutableMapLister   listers.MutableMapLister
	immutableMapLister listers.ImmutableMapLister
}

// Check that we implement the controller.Reconciler interface.
var _ controller.Reconciler = (*Reconciler)(nil)

func init() {
	// Add mutable-controller types to the default Kubernetes Scheme so Events can be
	// logged for mutable-controller types.
	boosscheme.AddToScheme(scheme.Scheme)
}

// NewController returns a new mutable controller
func NewController(
	opt reconciler.Options,
	boosclientset clientset.Interface,
	mutableMapInformer informers.MutableMapInformer,
	immutableMapInformer informers.ImmutableMapInformer,
) *controller.Impl {
	r := &Reconciler{
		Base:               reconciler.NewBase(opt, controllerAgentName),
		boosclientset:      boosclientset,
		mutableMapLister:   mutableMapInformer.Lister(),
		immutableMapLister: immutableMapInformer.Lister(),
	}
	impl := controller.NewImpl(r, r.Logger, "MutableMaps",
		reconciler.MustNewStatsReporter("MutableMaps", r.Logger))

	r.Logger.Info("Setting up event handlers")

	// Set up an event handler for when MutableMap resources change.
	mutableMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    impl.Enqueue,
		UpdateFunc: controller.PassNew(impl.Enqueue),
		DeleteFunc: impl.Enqueue,
	})

	// Set up an event handler for when Knative Service resources that we own change.
	immutableMapInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("MutableMap")),
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
	original, err := c.mutableMapLister.MutableMaps(namespace).Get(name)
	if errors.IsNotFound(err) {
		// The MutableMap resource may no longer exist, in which case we stop processing.
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

func (c *Reconciler) reconcile(ctx context.Context, im *v1alpha1.MutableMap) error {
	if err := c.reconcileImmutableMap(ctx, im); err != nil {
		return err
	}
	return nil
}

func (c *Reconciler) reconcileImmutableMap(ctx context.Context, im *v1alpha1.MutableMap) error {
	cmName := names.ImmutableMap(im)
	cm, err := c.immutableMapLister.ImmutableMaps(im.Namespace).Get(cmName)
	if apierrs.IsNotFound(err) {
		desiredCM := resources.MakeImmutableMap(im)
		cm, err = c.boosclientset.BoosV1alpha1().ImmutableMaps(im.Namespace).Create(desiredCM)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		desiredCM := resources.MakeImmutableMap(im)
		if !equality.Semantic.DeepEqual(cm.Spec, desiredCM.Spec) {
			cm = cm.DeepCopy()
			cm.Spec = desiredCM.Spec
			cm, err = c.boosclientset.BoosV1alpha1().ImmutableMaps(im.Namespace).Update(cm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
