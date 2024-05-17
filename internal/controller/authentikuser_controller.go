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

// AuthentikUserReconciler reconciles a AuthentikUser object
type AuthentikUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuthentikUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *AuthentikUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// Fetch the AuthentikGroup instance
	authentikUser := &appsv1.AuthentikUser{}
	err := r.Get(ctx, req.NamespacedName, authentikUser)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("AuthentikUser resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get AuthentikUser.")
		return ctrl.Result{}, err
	}

	// Check if the AuthentikGroup instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isAuthentikUserMarkedToBeDeleted := authentikUser.GetDeletionTimestamp() != nil

	if isAuthentikUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(authentikUser, authentikFinalizer) {
			if err := r.finalizeAuthentikUser(ctx, reqLogger, authentikUser); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(authentikUser, authentikFinalizer)
			err := r.Update(ctx, authentikUser)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else {
		// The deletion timestamp is not set, so create/update the resource in Authentik
		if err := r.createOrUpdateAuthentikUser(ctx, reqLogger, authentikUser); err != nil {
			return ctrl.Result{}, err
		}

		reqLogger.Info("Processed user", "userName", authentikUser.Spec.Username)
	}

	// Add the finalizer to any CRD that does not have it yet
	if !controllerutil.ContainsFinalizer(authentikUser, authentikFinalizer) {
		controllerutil.AddFinalizer(authentikUser, authentikFinalizer)
		err := r.Update(ctx, authentikUser)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AuthentikUserReconciler) finalizeAuthentikUser(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikUser) error {
	cl := authentik.GetClient(ctx)

	err := authentik.DeleteUser(&cl, m.Spec.Name)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully deleted AuthentikUser")
	return nil
}

func (r *AuthentikUserReconciler) createOrUpdateAuthentikUser(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikUser) error {
	user := api.User{
		Name:     m.Spec.Name,
		Username: m.Spec.Username,
		Email:    &m.Spec.Email,
		Groups:   m.Spec.Groups,
	}

	cl := authentik.GetClient(ctx)

	newUser, err := authentik.CreateUser(&cl, &user)

	if err != nil {
		return err
	}

	err = authentik.SynchronizeGroups(&cl, newUser, m.Spec.Groups)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully created/updated AuthentikUser")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthentikUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.AuthentikUser{}).
		Complete(r)
}
