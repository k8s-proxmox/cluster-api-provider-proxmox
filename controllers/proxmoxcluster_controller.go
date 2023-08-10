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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util"
	capiannotations "sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/storage"
)

// ProxmoxClusterReconciler reconciles a ProxmoxCluster object
type ProxmoxClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=proxmoxclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

func (r *ProxmoxClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := log.FromContext(ctx)

	proxmoxCluster := &infrav1.ProxmoxCluster{}
	if err := r.Get(ctx, req.NamespacedName, proxmoxCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("ProxmoxCluster \"%s\" is not found or already deleted", proxmoxCluster.Name))
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch ProxmoxCluster resource")
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, proxmoxCluster.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to get owner cluster")
		return ctrl.Result{}, err
	}

	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if capiannotations.IsPaused(cluster, proxmoxCluster) {
		log.Info("ProxmoxCluster or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	// Create the scope
	clusterScope, err := scope.NewClusterScope(ctx, scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		ProxmoxCluster: proxmoxCluster,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	if !proxmoxCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope)
	}

	return r.reconcile(ctx, clusterScope)
}

func (r *ProxmoxClusterReconciler) reconcile(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling ProxmoxCluster")

	if ok := controllerutil.AddFinalizer(clusterScope.ProxmoxCluster, infrav1.ClusterFinalizer); ok {
		log.Info("update finalizer to ProxmoxCluster")
	}
	if err := clusterScope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	reconcilers := []cloud.Reconciler{
		storage.NewService(clusterScope),
	}

	for _, r := range reconcilers {
		if err := r.Reconcile(ctx); err != nil {
			log.Error(err, "Reconcile error")
			record.Warnf(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Reconcile error - %v", err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	controlPlaneEndpoint := clusterScope.ControlPlaneEndpoint()
	if controlPlaneEndpoint.Host == "" {
		log.Info("ProxmoxCluster does not have control-plane endpoint yet. Reconciling")
		record.Event(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Waiting for control-plane endpoint")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	log.Info("Reconciled ProxmoxCluster")
	record.Eventf(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Got control-plane endpoint - %s", controlPlaneEndpoint.Host)
	clusterScope.SetReady()
	record.Event(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

func (r *ProxmoxClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling Delete ProxmoxCluster")

	reconcilers := []cloud.Reconciler{
		storage.NewService(clusterScope),
	}

	for _, r := range reconcilers {
		if err := r.Delete(ctx); err != nil {
			log.Error(err, "Reconcile error")
			record.Warnf(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Reconcile error - %v", err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	log.Info("Reconciled ProxmoxCluster")
	controllerutil.RemoveFinalizer(clusterScope.ProxmoxCluster, infrav1.ClusterFinalizer)
	record.Event(clusterScope.ProxmoxCluster, "ProxmoxClusterReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProxmoxClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.ProxmoxCluster{}).
		Complete(r)
}
