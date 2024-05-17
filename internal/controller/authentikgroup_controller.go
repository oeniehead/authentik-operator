/*
Copyright 2024.

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

package controller

import (
	"context"
	"github.com/go-logr/logr"
	"goauthentik.io/api/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/oeniehead/authentik-operator/api/v1"
	authentik "github.com/oeniehead/authentik-operator/internal/api"
)

const authentikFinalizer = "oeniehead.net/finalizer"

// AuthentikGroupReconciler reconciles a AuthentikGroup object
type AuthentikGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikgroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikgroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikgroups/finalizers,verbs=update

func (r *AuthentikGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// Fetch the AuthentikGroup instance
	authentikGroup := &appsv1.AuthentikGroup{}
	err := r.Get(ctx, req.NamespacedName, authentikGroup)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("AuthentikGroup resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get AuthentikGroup.")
		return ctrl.Result{}, err
	}

	// Check if the AuthentikGroup instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isAuthentikGroupMarkedToBeDeleted := authentikGroup.GetDeletionTimestamp() != nil

	if isAuthentikGroupMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(authentikGroup, authentikFinalizer) {
			if err := r.finalizeAuthentikGroup(ctx, reqLogger, authentikGroup); err != nil {
				return ctrl.Result{}, err
			}

			// Remove memcachedFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(authentikGroup, authentikFinalizer)
			err := r.Update(ctx, authentikGroup)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else {
		// The deletion timestamp is not set, so create/update the resource in Authentik
		if err := r.createOrUpdateAuthentikGroup(ctx, reqLogger, authentikGroup); err != nil {
			return ctrl.Result{}, err
		}

		reqLogger.Info("Processed group", "groupName", authentikGroup.Name)
	}

	// Add the finalizer to any CRD that does not have it yet
	if !controllerutil.ContainsFinalizer(authentikGroup, authentikFinalizer) {
		controllerutil.AddFinalizer(authentikGroup, authentikFinalizer)
		err := r.Update(ctx, authentikGroup)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AuthentikGroupReconciler) finalizeAuthentikGroup(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikGroup) error {
	cl := authentik.GetClient(ctx)

	err := authentik.DeleteGroup(&cl, m.Spec.Name)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully deleted AuthentikGroup")
	return nil
}

func (r *AuthentikGroupReconciler) createOrUpdateAuthentikGroup(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikGroup) error {
	parent := api.NewNullableString(m.Spec.Parent)

	group := api.Group{
		Name:        m.Spec.Name,
		IsSuperuser: &m.Spec.IsAdmin,
		Parent:      *parent,
	}

	cl := authentik.GetClient(ctx)

	_, err := authentik.CreateGroup(&cl, &group)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully created AuthentikGroup")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthentikGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.AuthentikGroup{}).
		Complete(r)
}
