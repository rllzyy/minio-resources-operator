package miniouser

import (
	"context"
	"fmt"

	"github.com/minio/minio/pkg/madmin"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	"github.com/Walkbase/minio-resources-operator/pkg/vault"
)

var log = logf.Log.WithName("controller_miniouser")

const minioUserFinalizer = "finalizer.user.minio.walkbase.com"

// Add creates a new MinioUser Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMinioUser{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("miniouser-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("controller.New: %w", err)
	}

	// Watch for changes to primary resource MinioUser
	err = c.Watch(&source.Kind{Type: &miniov1alpha1.MinioUser{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return fmt.Errorf("c.Watch: %w", err)
	}

	return nil
}

// blank assignment to verify that ReconcileMinioUser implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMinioUser{}

// ReconcileMinioUser reconciles a MinioUser object
type ReconcileMinioUser struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MinioUser object and makes changes based on the state read
// and what is in the MinioUser.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMinioUser) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MinioUser")

	// Fetch the MinioUser instance
	instance := &miniov1alpha1.MinioUser{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, fmt.Errorf("r.client.Get: %w", err)
	}

	minioServer := &miniov1alpha1.MinioServer{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{
		Name: instance.Spec.Server,
	}, minioServer); err != nil {
		return reconcile.Result{}, fmt.Errorf("r.client.Get: %w", err)
	}

	// doc is https://github.com/minio/minio/tree/master/pkg/madmin
	minioAdminClient, err := madmin.New(minioServer.Spec.GetHostname(), minioServer.Spec.AccessKey, minioServer.Spec.SecretKey, minioServer.Spec.SSL)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("madmin.New: %w", err)
	}

	vaultCreds, err := vault.GetCredentials(instance.Name)
	reqLogger.Info("Got user credentials")

	reqLogger.Info("List all Minio users")
	allUsers, err := minioAdminClient.ListUsers()
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("minioAdminClient.ListUsers: %w", err)
	}
	reqLogger.Info("Got user list")
	existingUser, isUserExists := allUsers[vaultCreds.AccessKey]

	userPolicyName := fmt.Sprintf("_generator_%s", vaultCreds.AccessKey)
	reqLogger = reqLogger.WithValues("Minio.Policy", userPolicyName)

	reqLogger.Info("List all Minio policies")
	allPolicies, err := minioAdminClient.ListCannedPolicies()
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
				if err = minioAdminClient.RemoveUser(vaultCreds.AccessKey); err != nil {
					return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveUser: %w", err)
				}
				reqLogger.Info("Minio user removed")
			} else {
				reqLogger.Info("Minio user already removed")
			}

			if isPolicyExists {
				reqLogger.Info("Delete Minio canned policy")
				if err = minioAdminClient.RemoveCannedPolicy(userPolicyName); err != nil {
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
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
			}
			reqLogger.Info("Finalizer deleted")
		} else {
			reqLogger.Info("Instance marked for deletion, but not minioUserFinalizer")
		}
		return reconcile.Result{}, nil
	}

	if err := controllerutil.SetControllerReference(minioServer, instance, r.scheme); err != nil {
		return reconcile.Result{}, fmt.Errorf("controllerutil.SetControllerReference: %w", err)
	}

	if !finalizerPresent {
		reqLogger.Info("No finalizer, add it")
		instance.SetFinalizers(append(instance.GetFinalizers(), minioUserFinalizer))
		if err = r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("r.client.Update: %w", err)
		}
		reqLogger.Info("Finalizer added")
	}

	isUserPolicy := len(instance.Spec.Policy) != 0
	needCreate := true
	if isPolicyExists {
		if !isUserPolicy {
			reqLogger.Info("Policy exists but unused, remove")
			if err = minioAdminClient.RemoveCannedPolicy(userPolicyName); err != nil {
				return reconcile.Result{}, fmt.Errorf("minioAdminClient.RemoveCannedPolicy: %w", err)
			}
			needCreate = false
			reqLogger.Info("Unused policy removed")
		} else {
			reqLogger.Info("Policy already exists, check if update needed")
			if existingPolicy != instance.Spec.Policy {
				reqLogger.Info("Policy key is different, recreate")
				reqLogger.Info("Delete existing policy")
				if err = minioAdminClient.RemoveCannedPolicy(userPolicyName); err != nil {
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
		if err = minioAdminClient.AddCannedPolicy(userPolicyName, instance.Spec.Policy); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.AddCannedPolicy: %w", err)
		}
		reqLogger.Info("New policy created")
	}

	if !isUserExists {
		reqLogger.Info("Create user")
		if err = minioAdminClient.AddUser(vaultCreds.AccessKey, vaultCreds.SecretKey); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.AddUser: %w", err)
		}
		reqLogger.Info("User created")
	}

	if isUserPolicy && (existingUser.PolicyName != userPolicyName || needCreate) {
		reqLogger.Info("Set user policy")
		if err = minioAdminClient.SetPolicy(userPolicyName, vaultCreds.AccessKey, false); err != nil {
			return reconcile.Result{}, fmt.Errorf("minioAdminClient.SetPolicy: %w", err)
		}
		reqLogger.Info("User policy set")
	}

	reqLogger.Info("Set user secret key")
	if err = minioAdminClient.SetUser(vaultCreds.AccessKey, vaultCreds.SecretKey, madmin.AccountEnabled); err != nil {
		return reconcile.Result{}, fmt.Errorf("minioAdminClient.SetUser: %w", err)
	}
	reqLogger.Info("Secret key set, reconcilied")

	return reconcile.Result{}, nil
}
