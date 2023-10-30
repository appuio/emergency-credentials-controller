package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ServiceAccountReconciler reconciles a ClusterVersion object
type ServiceAccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch

// Reconcile reconciles ServiceAccounts.
func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("ServiceAccountReconciler.Reconcile")
	l.Info("Reconciling ServiceAccount")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// The controller is setup to watch the managed ServiceAccount object.
func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}
