/*
Copyright 2018 The Knative Authors

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

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/knative/pkg/logging"
	"github.com/knative/pkg/signals"
	"github.com/knative/pkg/system"
	"github.com/knative/pkg/version"
	"github.com/knative/pkg/webhook"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
	clientset "github.com/mattmoor/boo-maps/pkg/client/clientset/versioned"
	informers "github.com/mattmoor/boo-maps/pkg/client/informers/externalversions"
)

const (
	component = "webhook"
)

var (
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()
	logger := logging.FromContext(context.TODO()).Named("controller")

	logger.Info("Starting the Configuration Webhook")

	// Set up signals so we handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	clusterConfig, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		logger.Fatalw("Failed to get cluster config", zap.Error(err))
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatalw("Failed to get the client set", zap.Error(err))
	}

	boosclient, err := clientset.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatalf("Error building serving clientset: %v", err)
	}

	if err := version.CheckMinimumVersion(kubeClient.Discovery()); err != nil {
		logger.Fatalf("Version check failed: %v", err)
	}

	boosInformerFactory := informers.NewSharedInformerFactory(boosclient, 10*time.Hour)

	mutableMapInformer := boosInformerFactory.Boos().V1alpha1().MutableMaps()

	go mutableMapInformer.Informer().Run(stopCh)

	// Wait for the caches to be synced before starting controllers.
	logger.Info("Waiting for informer caches to sync")
	for i, synced := range []cache.InformerSynced{
		mutableMapInformer.Informer().HasSynced,
	} {
		if ok := cache.WaitForCacheSync(stopCh, synced); !ok {
			logger.Fatalf("failed to wait for cache at index %v to sync", i)
		}
	}

	ml := mutableMapInformer.Lister()

	v1alpha1.FreezeConfigMap = func(namespace, name string) string {
		logger.Infof("Asked to freeze: %s", name)
		mm, err := ml.MutableMaps(namespace).Get(name)
		if errors.IsNotFound(err) {
			return name
		}
		return fmt.Sprintf("%s-%05d", mm.Name, mm.Generation)
	}

	options := webhook.ControllerOptions{
		ServiceName:    "webhook",
		DeploymentName: "webhook",
		Namespace:      system.Namespace(),
		Port:           443,
		SecretName:     "webhook-certs",
		WebhookName:    "webhook.serving.knative.dev",
	}
	controller := webhook.AdmissionController{
		Client:  kubeClient,
		Options: options,
		Handlers: map[schema.GroupVersionKind]webhook.GenericCRD{
			v1alpha1.SchemeGroupVersion.WithKind("ImmutableMap"): &v1alpha1.ImmutableMap{},
			v1alpha1.SchemeGroupVersion.WithKind("MutableMap"):   &v1alpha1.MutableMap{},
			appsv1.SchemeGroupVersion.WithKind("Deployment"):     &v1alpha1.WithPod{},
			appsv1.SchemeGroupVersion.WithKind("ReplicaSet"):     &v1alpha1.WithPod{},
			appsv1.SchemeGroupVersion.WithKind("StatefulSet"):    &v1alpha1.WithPod{},
			appsv1.SchemeGroupVersion.WithKind("DaemonSet"):      &v1alpha1.WithPod{},
			batchv1.SchemeGroupVersion.WithKind("Job"):           &v1alpha1.WithPod{},
		},
		Logger: logger,
	}
	if err = controller.Run(stopCh); err != nil {
		logger.Fatalw("Failed to start the admission controller", zap.Error(err))
	}
}
