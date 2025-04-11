/*
Copyright 2025 Ciprian Andrei.

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

	"cloud.google.com/go/storage"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mygroupv1 "github.com/andreistefanciprian/cloud-storage-controller/api/v1"
)

// CloudBucketReconciler reconciles a CloudBucket object
type CloudBucketReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	GCSClient *storage.Client
}

//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CloudBucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling CloudBucket", "namespace", req.Namespace, "name", req.Name)
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudBucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.CloudBucket{}).
		Complete(r)
}
