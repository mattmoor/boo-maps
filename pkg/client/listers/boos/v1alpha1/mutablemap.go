/*
Copyright 2019 Matt Moore

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/mattmoor/boo-maps/pkg/apis/boos/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MutableMapLister helps list MutableMaps.
type MutableMapLister interface {
	// List lists all MutableMaps in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.MutableMap, err error)
	// MutableMaps returns an object that can list and get MutableMaps.
	MutableMaps(namespace string) MutableMapNamespaceLister
	MutableMapListerExpansion
}

// mutableMapLister implements the MutableMapLister interface.
type mutableMapLister struct {
	indexer cache.Indexer
}

// NewMutableMapLister returns a new MutableMapLister.
func NewMutableMapLister(indexer cache.Indexer) MutableMapLister {
	return &mutableMapLister{indexer: indexer}
}

// List lists all MutableMaps in the indexer.
func (s *mutableMapLister) List(selector labels.Selector) (ret []*v1alpha1.MutableMap, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MutableMap))
	})
	return ret, err
}

// MutableMaps returns an object that can list and get MutableMaps.
func (s *mutableMapLister) MutableMaps(namespace string) MutableMapNamespaceLister {
	return mutableMapNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MutableMapNamespaceLister helps list and get MutableMaps.
type MutableMapNamespaceLister interface {
	// List lists all MutableMaps in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.MutableMap, err error)
	// Get retrieves the MutableMap from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.MutableMap, error)
	MutableMapNamespaceListerExpansion
}

// mutableMapNamespaceLister implements the MutableMapNamespaceLister
// interface.
type mutableMapNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MutableMaps in the indexer for a given namespace.
func (s mutableMapNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MutableMap, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MutableMap))
	})
	return ret, err
}

// Get retrieves the MutableMap from the indexer for a given namespace and name.
func (s mutableMapNamespaceLister) Get(name string) (*v1alpha1.MutableMap, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mutablemap"), name)
	}
	return obj.(*v1alpha1.MutableMap), nil
}
