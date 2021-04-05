// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type endpointReconciler struct {
	client       client.Client
	eventHandler cache.ResourceEventHandler
	logrus.FieldLogger
}

// NewEndpointController creates the endpoint controller from mgr. The controller will be pre-configured
// to watch for Endpoint objects across all namespaces.
func NewEndpointController(mgr manager.Manager, eventHandler cache.ResourceEventHandler, log logrus.FieldLogger) (controller.Controller, error) {
	r := &endpointReconciler{
		client:       mgr.GetClient(),
		eventHandler: eventHandler,
		FieldLogger:  log,
	}
	c, err := controller.New("endpoint-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}
	if err := c.Watch(&source.Kind{Type: &corev1.Endpoints{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *endpointReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	// Fetch the Service from the cache.
	endpoint := &corev1.Endpoints{}
	err := r.client.Get(ctx, request.NamespacedName, endpoint)
	if errors.IsNotFound(err) {
		r.Error(nil, "Could not find Endpoints %q in Namespace %q", request.Name, request.Namespace)
		return reconcile.Result{}, nil
	}

	// Check if object is deleted.
	if !endpoint.ObjectMeta.DeletionTimestamp.IsZero() {
		r.eventHandler.OnDelete(endpoint)
		return reconcile.Result{}, nil
	}

	// Pass the new changed object off to the eventHandler.
	r.eventHandler.OnAdd(endpoint)

	return reconcile.Result{}, nil
}
