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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/oeniehead/authentik-operator/api/v1"
)

// AuthentikApplicationReconciler reconciles a AuthentikApplication object
type AuthentikApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikapplications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.oeniehead.net,resources=authentikapplications/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuthentikApplication object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *AuthentikApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// Fetch the AuthentikGroup instance
	authentikApplication := &appsv1.AuthentikApplication{}
	err := r.Get(ctx, req.NamespacedName, authentikApplication)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("AuthentikApplication resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get AuthentikApplication.")
		return ctrl.Result{}, err
	}

	// Check if the AuthentikApplication instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isAuthentikUserMarkedToBeDeleted := authentikApplication.GetDeletionTimestamp() != nil

	if isAuthentikUserMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(authentikApplication, authentikFinalizer) {
			if err := r.finalizeAuthentikApplication(ctx, reqLogger, authentikApplication); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(authentikApplication, authentikFinalizer)
			err := r.Update(ctx, authentikApplication)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	} else {
		// The deletion timestamp is not set, so create/update the resource in Authentik
		if err := r.createOrUpdateAuthentikApplication(ctx, reqLogger, authentikApplication); err != nil {
			return ctrl.Result{}, err
		}

		reqLogger.Info("Processed application", "userName", authentikApplication.Spec.Name)
	}

	// Add the finalizer to any CRD that does not have it yet
	if !controllerutil.ContainsFinalizer(authentikApplication, authentikFinalizer) {
		controllerutil.AddFinalizer(authentikApplication, authentikFinalizer)
		err := r.Update(ctx, authentikApplication)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AuthentikApplicationReconciler) finalizeAuthentikApplication(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikApplication) error {
	cl := authentik.GetClient(ctx)

	err := authentik.DeleteApplication(&cl, m.Spec.Slug)

	if err != nil {
		return err
	}

	reqLogger.Info("Successfully deleted AuthentikApplication")
	return nil
}

func (r *AuthentikApplicationReconciler) createOrUpdateAuthentikApplication(ctx context.Context, reqLogger logr.Logger, m *appsv1.AuthentikApplication) error {
	cl := authentik.GetClient(ctx)

	existingApplication, _ := authentik.GetApplication(&cl, m.Spec.Slug)

	existingProvider, err := authentik.GetProvider(&cl, m.Spec.Provider)
	if err != nil {
		return err
	}
	if existingProvider == nil {
		return fmt.Errorf("provider %s not found", m.Spec.Provider)
	}

	if existingApplication == nil {
		application := api.Application{
			Name:     m.Spec.Name,
			Slug:     m.Spec.Slug,
			Group:    &m.Spec.Group,
			Provider: *api.NewNullableInt32(&existingProvider.Pk),
		}

		existingApplication, err = authentik.CreateApplication(&cl, &application)

		if err != nil {
			return err
		}
	}

	for _, groupName := range m.Spec.UserGroups {
		group, err := authentik.GetGroup(&cl, groupName)
		if err != nil {
			return err
		}
		if group == nil {
			return fmt.Errorf("group %s not found", groupName)
		}

		binding, err := authentik.GetGroupBinding(&cl, existingApplication.Pk, group.Pk)
		if err != nil {
			return err
		}

		if binding == nil {
			err := authentik.BindApplicationToGroup(&cl, existingApplication.Pk, group.Pk)
			if err != nil {
				return err
			}
		}
	}

	secret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: m.Spec.SecretName, Namespace: m.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		secret, err = r.defineSecret(m.Spec.SecretName, m.Namespace, *existingProvider.ClientId, *existingProvider.ClientSecret, m)
		if err != nil {
			return err
		}
		err := r.Create(ctx, secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	reqLogger.Info("Successfully created AuthentikApplication")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthentikApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.AuthentikApplication{}).
		Complete(r)
}

func (r *AuthentikApplicationReconciler) defineSecret(name string, namespace string, clientId string, clientSecret string, application *appsv1.AuthentikApplication) (*corev1.Secret, error) {
	secret := make(map[string]string)
	secret["OAUTH_CLIENT_ID"] = clientId
	secret["OAUTH_CLIENT_SECRET"] = clientSecret

	sec := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Immutable:  new(bool),
		Data:       map[string][]byte{},
		StringData: secret,
		Type:       "Opaque",
	}

	// Used to ensure that the secret will be deleted when the custom resource object is removed
	ctrl.SetControllerReference(application, sec, r.Scheme)

	return sec, nil
}
