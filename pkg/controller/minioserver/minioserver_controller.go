package minioserver

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	miniov1alpha1 "github.com/robotinfra/minio-resources-operator/pkg/apis/minio/v1alpha1"
	"github.com/robotinfra/minio-resources-operator/pkg/utils"
)

var (
	log     = logf.Log.WithName("controller_minioserver")
	servers sync.Map
)

const minioServerFinalizer = "finalizer.server.minio.robotinfra.com"

// GetMinioServer if available, return a MinioServer that CR is present
func GetMinioServer(name string) *miniov1alpha1.MinioServer {
	server, found := servers.Load(name)
	if !found {
		return nil
	}
	return server.(*miniov1alpha1.MinioServer)
}

// Add creates a new MinioServer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMinioServer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("minioserver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("controller.New: %w", err)
	}

	// Watch for changes to primary resource MinioUser
	err = c.Watch(&source.Kind{Type: &miniov1alpha1.MinioServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return fmt.Errorf("c.Watch: %w", err)
	}

	return nil
}

// blank assignment to verify that ReconcileMinioServer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMinioServer{}

// ReconcileMinioServer reconciles a MinioServer object
type ReconcileMinioServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MinioServer object and makes changes based on the state read
// and what is in the MinioServer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMinioServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MinioServer")

	// Fetch the MinioServer instance
	instance := &miniov1alpha1.MinioServer{}
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

	finalizerPresent := utils.Contains(instance.GetFinalizers(), minioServerFinalizer)

	if instance.GetDeletionTimestamp() != nil {
		if finalizerPresent {
			servers.Delete(instance.ObjectMeta.Name)

			// Remove minioServerFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			reqLogger.Info("Removed server, delete finalizer")
			instance.SetFinalizers(utils.Remove(instance.GetFinalizers(), minioServerFinalizer))
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
			}
			reqLogger.Info("Finalizer deleted")
		} else {
			reqLogger.Info("Instance marked for deletion, but not minioServerFinalizer")
		}
		return reconcile.Result{}, nil
	}

	servers.Store(instance.ObjectMeta.Name, &instance)

	if !finalizerPresent {
		reqLogger.Info("No finalizer, add it")
		instance.SetFinalizers(append(instance.GetFinalizers(), minioServerFinalizer))
		if err = r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
		}
		reqLogger.Info("Finalizer added")
	}

	return reconcile.Result{}, nil
}
