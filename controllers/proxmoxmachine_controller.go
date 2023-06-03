/*
Copyright 2023 Teppei Sudo.

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
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute"
)

// ProxmoxMachineReconciler reconciles a ProxmoxMachine object
type ProxmoxMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ProxmoxMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)

	proxmoxMachine := &infrav1.ProxmoxMachine{}
	err := r.Get(ctx, req.NamespacedName, proxmoxMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	machine, err := util.GetOwnerMachine(ctx, r.Client, proxmoxMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("machine", machine.Name)
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("Machine is missing cluster label or cluster does not exist")

		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, proxmoxMachine) {
		log.Info("ProxmoxMachine or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)
	proxmoxCluster := &infrav1.ProxmoxCluster{}
	proxmoxClusterKey := client.ObjectKey{
		Namespace: proxmoxMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, proxmoxClusterKey, proxmoxCluster); err != nil {
		log.Info("ProxmoxCluster is not available yet")
		return ctrl.Result{}, nil
	}

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(ctx, scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		ProxmoxCluster: proxmoxCluster,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create the machine scope
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		Client:         r.Client,
		Machine:        machine,
		ProxmoxMachine: proxmoxMachine,
		ClusterGetter:  clusterScope,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any ProxmoxMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted machines
	if !proxmoxMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineScope)
	}

	// Handle non-deleted machines
	return r.reconcile(ctx, machineScope)
}

func (r *ProxmoxMachineReconciler) reconcile(ctx context.Context, machineScope *scope.MachineScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling ProxmoxMachine")

	if ok := controllerutil.AddFinalizer(machineScope.ProxmoxMachine, infrav1.MachineFinalizer); ok {
		log.Info("update finalizer to ProxmoxMachine")
	}

	reconcilers := []cloud.Reconciler{
		compute.NewService(machineScope),
	}

	for _, r := range reconcilers {
		if err := r.Reconcile(ctx); err != nil {
			log.Error(err, "Reconcile error")
			record.Warnf(machineScope.ProxmoxMachine, "ProxmoxMachineReconcile", "Reconcile error - %v", err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	instanceState := *machineScope.GetInstanceStatus()
	switch instanceState {
	case infrav1.InstanceStatusRunning:
		log.Info("ProxmoxMachine instance is running", "instance-id", *machineScope.GetInstanceID())
		record.Eventf(machineScope.ProxmoxMachine, "ProxmoxMachineReconcile", "ProxmoxMachine instance is running - instance-id: %s", *machineScope.GetInstanceID())
		record.Event(machineScope.ProxmoxMachine, "ProxmoxMachineReconcile", "Reconciled")
		machineScope.SetReady()
		return ctrl.Result{}, nil
	default:
		machineScope.SetFailureReason(capierrors.UpdateMachineError)
		machineScope.SetFailureMessage(errors.Errorf("ProxmoxMachine instance state %s is unexpected", instanceState))
		return ctrl.Result{Requeue: true}, nil
	}

}

func (r *ProxmoxMachineReconciler) reconcileDelete(ctx context.Context, machineScope *scope.MachineScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Delete ProxmoxMachine")

	reconcilers := []cloud.Reconciler{
		compute.NewService(machineScope),
	}

	for _, r := range reconcilers {
		if err := r.Delete(ctx); err != nil {
			log.Error(err, "Reconcile error")
			record.Warnf(machineScope.ProxmoxMachine, "ProxmoxMachineReconcile", "Reconcile error - %v", err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	controllerutil.RemoveFinalizer(machineScope.ProxmoxMachine, infrav1.MachineFinalizer)
	record.Event(machineScope.ProxmoxMachine, "ProxmoxMachineReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProxmoxMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ProxmoxMachine{}).
		Complete(r)
}
