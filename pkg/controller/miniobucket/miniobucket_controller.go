package miniobucket

import (
	"context"
	"fmt"

	"github.com/minio/minio-go"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	miniov1alpha1 "github.com/Walkbase/minio-resources-operator/pkg/apis/minio/v1alpha1"
	"github.com/Walkbase/minio-resources-operator/pkg/utils"
)

var log = logf.Log.WithName("controller_miniobucket")

const minioBucketFinalizer = "finalizer.bucket.minio.robotinfra.com"

// Add creates a new MinioBucket Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMinioBucket{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("miniobucket-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("controller.New: %w", err)
	}

	// Watch for changes to primary resource MinioBucket
	err = c.Watch(&source.Kind{Type: &miniov1alpha1.MinioBucket{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return fmt.Errorf("c.Watch: %w", err)
	}

	return nil
}

// blank assignment to verify that ReconcileMinioBucket implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMinioBucket{}

// ReconcileMinioBucket reconciles a MinioBucket object
type ReconcileMinioBucket struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MinioBucket object and makes changes based on the state read
// and what is in the MinioBucket.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMinioBucket) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MinioBucket")

	// Fetch the MinioBucket instance
	instance := &miniov1alpha1.MinioBucket{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	minioServer := &miniov1alpha1.MinioServer{}
	if err := r.client.Get(context.TODO(), client.ObjectKey{
		Name: instance.Spec.Server,
	}, minioServer); err != nil {
		return reconcile.Result{}, fmt.Errorf("r.client.Get: %w", err)
	}

	// doc is https://github.com/minio/minio/tree/master/pkg/madmin
	minioClient, err := minio.New(minioServer.Spec.GetHostname(), minioServer.Spec.AccessKey, minioServer.Spec.SecretKey, minioServer.Spec.SSL)
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
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
			}
			reqLogger.Info("Finalizer deleted")
		} else {
			reqLogger.Info("Instance marked for deletion, but not minioBucketFinalizer")
		}
		return reconcile.Result{}, nil
	}

	if err := controllerutil.SetControllerReference(minioServer, instance, r.scheme); err != nil {
		return reconcile.Result{}, fmt.Errorf("controllerutil.SetControllerReference: %w", err)
	}

	if !finalizerPresent {
		reqLogger.Info("No finalizer, add it")
		instance.SetFinalizers(append(instance.GetFinalizers(), minioBucketFinalizer))
		if err = r.client.Update(context.TODO(), instance); err != nil {
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
