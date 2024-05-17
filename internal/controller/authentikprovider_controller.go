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
	"fmt"
	"github.com/go-logr/logr"
	authentik "github.com/oeniehead/authentik-operator/internal/api"
	"goauthentik.io/api/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/oeniehead/authentik-operator/api/v1"
)

// AuthentikProviderReconciler reconciles a AuthentikProvider object
type AuthentikProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikproviders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikproviders/finalizers,verbs=update

func (r *AuthentikProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// Fetch the AuthentikGroup instance
	authentikProvider := &appsv1.AuthentikProvider{}
	err := r.Get(ctx, req.NamespacedName, authentikProvider)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("AuthentikProvider resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get AuthentikProvider.")
		return ctrl.Result{}, err
	}

	// Check if the AuthentikProvider instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isAuthentikProviderMarkedToBeDeleted := authentikProvider.GetDeletionTimestamp() != nil

	if isAuthentikProviderMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(authentikProvider, authentikFinalizer) {
			if err := r.finalizeAuthentikProvider(ctx, reqLogger, authentikProvider); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(authentikProvider, authentikFinalizer)
			err := r.Update(ctx, authentikProvider)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else {
		// The deletion timestamp is not set, so create/update the resource in Authentik
		if err := r.createOrUpdateAuthentikProvider(ctx, reqLogger, authentikProvider); err != nil {
			return ctrl.Result{}, err
		}

		reqLogger.Info("Processed user", "userName", authentikProvider.Spec.Name)
	}

	// Add the finalizer to any CRD that does not have it yet
	if !controllerutil.ContainsFinalizer(authentikProvider, authentikFinalizer) {
		controllerutil.AddFinalizer(authentikProvider, authentikFinalizer)
		err := r.Update(ctx, authentikProvider)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AuthentikProviderReconciler) finalizeAuthentikProvider(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikProvider) error {
	cl := authentik.GetClient(ctx)

	err := authentik.DeleteProvider(&cl, m.Spec.Name)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully deleted AuthentikProvider")
	return nil
}

func (r *AuthentikProviderReconciler) createOrUpdateAuthentikProvider(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikProvider) error {
	cl := authentik.GetClient(ctx)

	clientType, err := api.NewClientTypeEnumFromValue(m.Spec.ClientType)

	if err != nil {
		return err
	}

	var mappings []string

	for _, v := range m.Spec.ScopeMappings {
		mapping, err := authentik.GetScopeMapping(&cl, v)
		if err != nil {
			return err
		}
		if mapping == nil {
			return fmt.Errorf("scopemapping %s not found", v)
		}

		mappings = append(mappings, mapping.Pk)
	}

	authenticationFlow, err := authentik.GetFlow(&cl, m.Spec.AuthenticationFlow, "authentication")
	if err != nil {
		return fmt.Errorf("authentication flow %s not found", m.Spec.AuthenticationFlow)
	}
	if authenticationFlow == nil {
		return fmt.Errorf("authentication flow %s not found", m.Spec.AuthenticationFlow)
	}

	authorizationFlow, err := authentik.GetFlow(&cl, m.Spec.AuthorizationFlow, "authorization")
	if err != nil {
		return fmt.Errorf("authorization flow %s not found", m.Spec.AuthorizationFlow)
	}
	if authorizationFlow == nil {
		return fmt.Errorf("authorization flow %s not found", m.Spec.AuthorizationFlow)
	}

	existingProvider, err := authentik.GetProvider(&cl, m.Spec.Name)

	if err != nil {
		return err
	}

	if existingProvider != nil {
		return nil
	}

	provider := api.OAuth2Provider{
		Name:               m.Spec.Name,
		AuthenticationFlow: *api.NewNullableString(&authenticationFlow.Pk),
		AuthorizationFlow:  authorizationFlow.Pk,
		ClientType:         clientType,
		RedirectUris:       &m.Spec.RedirectUri,
		PropertyMappings:   mappings,
	}

	_, err = authentik.CreateProvider(&cl, &provider)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully created/updated AuthentikProvider")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthentikProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.AuthentikProvider{}).
		Complete(r)
}
