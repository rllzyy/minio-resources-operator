/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	miniov1 "github.com/Walkbase/minio-resources-operator/api/v1"
	"github.com/Walkbase/minio-resources-operator/utils"
	"github.com/Walkbase/minio-resources-operator/vault"
	madmin "github.com/minio/madmin-go"
)

// MinioUserReconciler reconciles a MinioUser object
type MinioUserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const minioUserFinalizer = "finalizer.user.minio.walkbase.com"

//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniousers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniousers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniousers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MinioUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MinioUserReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MinioUser")

	// Fetch the MinioUser instance
	instance := &miniov1.MinioUser{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, fmt.Errorf("r.Get: %w", err)
	}

	minioServer := &miniov1.MinioServer{}
	if err := r.Get(ctx, types.NamespacedName{
		Name: instance.Spec.Server,
	}, minioServer); err != nil {
		return reconcile.Result{}, fmt.Errorf("r.Get: %w", err)
	}

	serverCreds, err := vault.GetServerCredentials(minioServer.Name)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get server creds: %w", err)
	}

	// doc is https://github.com/minio/minio/tree/master/pkg/madmin
	minioAdminClient, err := madmin.New(minioServer.Spec.GetHostname(), serverCreds.AccessKey, serverCreds.SecretKey, minioServer.Spec.SSL)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("madmin.New: %w", err)
	}

	vaultCreds, err := vault.GetCredentials(instance.Name)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("vaultCreds: %w", err)
	}
	reqLogger.Info("Got user credentials")

	reqLogger.Info("List all Minio users")
	allUsers, err := minioAdminClient.ListUsers(context.Background())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("minioAdminClient.ListUsers: %w", err)
	}
	reqLogger.Info("Got user list")
	existingUser, isUserExists := allUsers[vaultCreds.AccessKey]

	userPolicyName := fmt.Sprintf("_generator_%s", vaultCreds.AccessKey)
	reqLogger = reqLogger.WithValues("Minio.Policy", userPolicyName)

	reqLogger.Info("List all Minio policies")
	allPolicies, err := minioAdminClient.ListCannedPolicies(context.Background())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("minioAdminClient.ListCannedPolicies: %w", err)
	}
	reqLogger.Info("Got policy list")
	existingPolicyBytes, isPolicyExists := allPolicies[userPolicyName]
	existingPolicy := string(existingPolicyBytes)

	finalizerPresent := utils.Contains(instance.GetFinalizers(), minioUserFinalizer)

	if instance.GetDeletionTimestamp() != nil {
		if finalizerPresent {
			// Run finalization logic for. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if isUserExists {
				reqLogger.Info("Instance marked for deletion, remove Minio user")
				if err = minioAdminClient.RemoveUser(context.Background(), vaultCreds.AccessKey); err != nil {
					return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveUser: %w", err)
				}
				reqLogger.Info("Minio user removed")
			} else {
				reqLogger.Info("Minio user already removed")
			}

			if isPolicyExists {
				reqLogger.Info("Delete Minio canned policy")
				if err = minioAdminClient.RemoveCannedPolicy(context.Background(), userPolicyName); err != nil {
					return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveCannedPolicy: %w", err)
				}
				reqLogger.Info("Minio policy removed")
			} else {
				reqLogger.Info("Minio policy already removed")
			}

			// Remove minioUserFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			reqLogger.Info("Delete finalizer")
			instance.SetFinalizers(utils.Remove(instance.GetFinalizers(), minioUserFinalizer))
			if err = r.Update(ctx, instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("r.Update: %w", err)
			}
			reqLogger.Info("Finalizer deleted")
		} else {
			reqLogger.Info("Instance marked for deletion, but not minioUserFinalizer")
		}
		return reconcile.Result{}, nil
	}

	if err := controllerutil.SetControllerReference(minioServer, instance, r.Scheme); err != nil {
		return reconcile.Result{}, fmt.Errorf("controllerutil.SetControllerReference: %w", err)
	}

	if !finalizerPresent {
		reqLogger.Info("No finalizer, add it")
		instance.SetFinalizers(append(instance.GetFinalizers(), minioUserFinalizer))
		if err = r.Update(ctx, instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("r.Update: %w", err)
		}
		reqLogger.Info("Finalizer added")
	}

	isUserPolicy := len(instance.Spec.Policy) != 0
	needCreate := true
	if isPolicyExists {
		if !isUserPolicy {
			reqLogger.Info("Policy exists but unused, remove")
			if err = minioAdminClient.RemoveCannedPolicy(context.Background(), userPolicyName); err != nil {
				return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveCannedPolicy: %w", err)
			}
			needCreate = false
			reqLogger.Info("Unused policy removed")
		} else {
			reqLogger.Info("Policy already exists, check if update needed")
			if existingPolicy != instance.Spec.Policy {
				reqLogger.Info("Policy key is different, recreate")
				reqLogger.Info("Delete existing policy")
				if err = minioAdminClient.RemoveCannedPolicy(context.Background(), userPolicyName); err != nil {
					return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveCannedPolicy: %w", err)
				}
				reqLogger.Info("Existing policy deleted")
			} else {
				needCreate = false
				reqLogger.Info("Policy is correct state")
			}
		}
	}

	if needCreate && isUserPolicy {
		reqLogger.Info("Create new policy")
		if err = minioAdminClient.AddCannedPolicy(context.Background(), userPolicyName, []byte(instance.Spec.Policy)); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.AddCannedPolicy: %w", err)
		}
		reqLogger.Info("New policy created")
	}

	if !isUserExists {
		reqLogger.Info("Create user")
		if err = minioAdminClient.AddUser(context.Background(), vaultCreds.AccessKey, vaultCreds.SecretKey); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.AddUser: %w", err)
		}
		reqLogger.Info("User created")
	}

	if isUserPolicy && (existingUser.PolicyName != userPolicyName || needCreate) {
		reqLogger.Info("Set user policy")
		if err = minioAdminClient.SetPolicy(context.Background(), userPolicyName, vaultCreds.AccessKey, false); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.SetPolicy: %w", err)
		}
		reqLogger.Info("User policy set")
	}

	reqLogger.Info("Set user secret key")
	if err = minioAdminClient.SetUser(context.Background(), vaultCreds.AccessKey, vaultCreds.SecretKey, madmin.AccountEnabled); err != nil {
		return reconcile.Result{}, fmt.Errorf("minioAdminClient.SetUser: %w", err)
	}
	reqLogger.Info("Secret key set, reconcilied")

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinioUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&miniov1.MinioUser{}).
		Complete(r)
}
