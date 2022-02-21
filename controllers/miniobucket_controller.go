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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	miniov1 "github.com/Walkbase/minio-resources-operator/api/v1"
	"github.com/Walkbase/minio-resources-operator/utils"
	"github.com/Walkbase/minio-resources-operator/vault"
	minio "github.com/minio/minio-go"
)

// MinioBucketReconciler reconciles a MinioBucket object
type MinioBucketReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const minioBucketFinalizer = "finalizer.bucket.minio.walkbase.com"

//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniobuckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniobuckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=minio.walkbase.com,resources=miniobuckets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MinioBucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MinioBucketReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MinioBucket")

	// Fetch the MinioBucket instance
	instance := &miniov1.MinioBucket{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, fmt.Errorf("r.client.Get: %w", err)
	}

	minioServer := &miniov1.MinioServer{}
	if err := r.Get(ctx, client.ObjectKey{
		Name: instance.Spec.Server,
	}, minioServer); err != nil {
		return reconcile.Result{}, fmt.Errorf("r.client.Get: %w", err)
	}

	serverCreds, err := vault.GetServerCredentials(minioServer.Name)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to get server creds: %w", err)
	}

	// doc is https://github.com/minio/minio/tree/master/pkg/madmin
	minioClient, err := minio.New(minioServer.Spec.GetHostname(), serverCreds.AccessKey, serverCreds.SecretKey, minioServer.Spec.SSL)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("minio.New: %w", err)
	}

	reqLogger.Info("Check if Minio bucket exists")
	bucketExist, err := minioClient.BucketExists(instance.Spec.Name)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("minioClient.BucketExists: %w", err)
	}
	reqLogger.Info("Got bucket info")

	finalizerPresent := utils.Contains(instance.GetFinalizers(), minioBucketFinalizer)

	if instance.GetDeletionTimestamp() != nil {
		if finalizerPresent {
			// Run finalization logic for. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if bucketExist {
				reqLogger.Info("Instance marked for deletion, remove Minio bucket")
				if err = minioClient.RemoveBucket(instance.Spec.Name); err != nil {
					return reconcile.Result{}, fmt.Errorf("minioClient.RemoveBucket: %w", err)
				}
				reqLogger.Info("Minio bucket removed")
			} else {
				reqLogger.Info("Minio bucket already removed")
			}

			// Remove minioUserFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			reqLogger.Info("Delete finalizer")
			instance.SetFinalizers(utils.Remove(instance.GetFinalizers(), minioBucketFinalizer))
			if err = r.Update(ctx, instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
			}
			reqLogger.Info("Finalizer deleted")
		} else {
			reqLogger.Info("Instance marked for deletion, but not minioBucketFinalizer")
		}
		return reconcile.Result{}, nil
	}

	if err := controllerutil.SetControllerReference(minioServer, instance, r.Scheme); err != nil {
		return reconcile.Result{}, fmt.Errorf("controllerutil.SetControllerReference: %w", err)
	}

	if !finalizerPresent {
		reqLogger.Info("No finalizer, add it")
		instance.SetFinalizers(append(instance.GetFinalizers(), minioBucketFinalizer))
		if err = r.Update(ctx, instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
		}
		reqLogger.Info("Finalizer added")
	}

	if bucketExist {
		reqLogger.Info("Get bucket policy")
		bucketPolicy, err := minioClient.GetBucketPolicy(instance.Spec.Name)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("minioClient.GetBucketPolicy: %w", err)
		}
		reqLogger.Info("Got bucket policy")

		if bucketPolicy != instance.Spec.Policy {
			reqLogger.Info("Bucket policy is different, replace", "Spec.Name", instance.Spec.Name)
			if err = minioClient.SetBucketPolicy(instance.Spec.Name, instance.Spec.Policy); err != nil {
				return reconcile.Result{}, fmt.Errorf("minioClient.SetBucketPolicy: %w", err)
			}
			reqLogger.Info("Bucket policy changed")
		} else {
			reqLogger.Info("Bucket policy is already correct")
		}
	} else {
		reqLogger.Info("Bucket don't exists, create")
		if err = minioClient.MakeBucket(instance.Spec.Name, ""); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioClient.MakeBucket: %w", err)
		}
		reqLogger.Info("Bucket created, set policy", "Spec.Name", instance.Spec.Name)
		if err = minioClient.SetBucketPolicy(instance.Spec.Name, instance.Spec.Policy); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioClient.SetBucketPolicy: %w", err)
		}
		reqLogger.Info("Bucket policy set")
	}

	reqLogger.Info("MinioBucket reconcilied")
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinioBucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&miniov1.MinioBucket{}).
		Complete(r)
}
