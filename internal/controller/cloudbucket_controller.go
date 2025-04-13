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
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mygroupv1 "github.com/andreistefanciprian/cloud-storage-controller/api/v1"
)

// CloudBucketReconciler reconciles a CloudBucket object
type CloudBucketReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	GCSClient     *storage.Client
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.example.com,resources=cloudbuckets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CloudBucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CloudBucket resource
	cloudBucket := &mygroupv1.CloudBucket{}
	err := r.Get(ctx, req.NamespacedName, cloudBucket)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("CloudBucket resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CloudBucket")
		ErrorsTotal.Inc()
		r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "FetchFailed", fmt.Sprintf("Failed to get CloudBucket: %v", err))
		return ctrl.Result{}, err
	}

	// Initialize status if empty
	if cloudBucket.Status.BucketExists == false && cloudBucket.Status.LastOperation == "" {
		cloudBucket.Status = mygroupv1.CloudBucketStatus{
			BucketExists:  false,
			LastOperation: "Pending",
		}
	}

	// Define finalizer
	const bucketFinalizer = "cloudbuckets.mygroup.example.com/finalizer"

	// Check if the CloudBucket is being deleted
	if cloudBucket.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(cloudBucket, bucketFinalizer) {
			if cloudBucket.Spec.DeletePolicy == "Delete" {
				log.Info("Deleting bucket due to CloudBucket deletion", "bucketName", cloudBucket.Spec.BucketName)
				err = r.deleteBucket(ctx, cloudBucket.Spec.BucketName)
				if err != nil {
					log.Error(err, "Failed to delete bucket")
					cloudBucket.Status.LastOperation = "Failed"
					cloudBucket.Status.ErrorMessage = err.Error()
					ErrorsTotal.Inc()
					r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "BucketFailed", fmt.Sprintf("Failed to delete bucket: %v", err))
					if updateErr := r.Status().Update(ctx, cloudBucket); updateErr != nil {
						log.Error(updateErr, "Failed to update CloudBucket status")
						ErrorsTotal.Inc()
					}
					return ctrl.Result{RequeueAfter: 30 * time.Second}, err
				}
				cloudBucket.Status.BucketExists = false
				cloudBucket.Status.LastOperation = "Deleted"
				cloudBucket.Status.ErrorMessage = ""
				BucketsDeleted.Inc()
				r.EventRecorder.Event(cloudBucket, corev1.EventTypeNormal, "BucketDeleted", "Bucket deleted successfully")
			} else {
				log.Info("Orphaning bucket due to deletePolicy", "bucketName", cloudBucket.Spec.BucketName)
				cloudBucket.Status.LastOperation = "Orphaned"
				cloudBucket.Status.ErrorMessage = ""
				BucketsOrphaned.Inc()
				r.EventRecorder.Event(cloudBucket, corev1.EventTypeNormal, "BucketOrphaned", "Bucket orphaned due to delete policy")
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(cloudBucket, bucketFinalizer)
			if err := r.Update(ctx, cloudBucket); err != nil {
				log.Error(err, "Failed to remove finalizer")
				ErrorsTotal.Inc()
				r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "FinalizerFailed", fmt.Sprintf("Failed to remove finalizer: %v", err))
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(cloudBucket, bucketFinalizer) {
		controllerutil.AddFinalizer(cloudBucket, bucketFinalizer)
		if err := r.Update(ctx, cloudBucket); err != nil {
			log.Error(err, "Failed to add finalizer")
			ErrorsTotal.Inc()
			r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "FinalizerFailed", fmt.Sprintf("Failed to add finalizer: %v", err))
			return ctrl.Result{}, err
		}
	}

	// Check if bucket exists
	exists, err := r.bucketExists(ctx, cloudBucket.Spec.BucketName)
	if err != nil {
		log.Error(err, "Failed to check bucket existence")
		cloudBucket.Status.LastOperation = "Failed"
		cloudBucket.Status.ErrorMessage = err.Error()
		ErrorsTotal.Inc()
		r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "BucketFailed", fmt.Sprintf("Failed to check bucket existence: %v", err))
		if updateErr := r.Status().Update(ctx, cloudBucket); updateErr != nil {
			log.Error(updateErr, "Failed to update CloudBucket status")
			ErrorsTotal.Inc()
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// If bucket doesn't exist, create it
	if !exists {
		log.Info("Creating bucket", "bucketName", cloudBucket.Spec.BucketName, "location", cloudBucket.Spec.Location)
		err = r.createBucket(ctx, cloudBucket.Spec.ProjectID, cloudBucket.Spec.BucketName, cloudBucket.Spec.Location)
		if err != nil {
			log.Error(err, "Failed to create bucket")
			cloudBucket.Status.BucketExists = false
			cloudBucket.Status.LastOperation = "Failed"
			cloudBucket.Status.ErrorMessage = err.Error()
			ErrorsTotal.Inc()
			r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "BucketFailed", fmt.Sprintf("Failed to create bucket: %v", err))
			if updateErr := r.Status().Update(ctx, cloudBucket); updateErr != nil {
				log.Error(updateErr, "Failed to update CloudBucket status")
				ErrorsTotal.Inc()
			}
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
		cloudBucket.Status.BucketExists = true
		if cloudBucket.Status.LastOperation == "Exists" || cloudBucket.Status.LastOperation == "Created" {
			cloudBucket.Status.LastOperation = "Recreated"
			BucketsRecreated.Inc()
			r.EventRecorder.Event(cloudBucket, corev1.EventTypeNormal, "BucketRecreated", "Bucket recreated after being missing")
		} else {
			cloudBucket.Status.LastOperation = "Created"
			BucketsCreated.Inc()
			r.EventRecorder.Event(cloudBucket, corev1.EventTypeNormal, "BucketCreated", "Bucket created successfully")
		}
		cloudBucket.Status.ErrorMessage = ""
	} else {
		// add event Bucket already exists only if it was not created by this operator
		if cloudBucket.Status.LastOperation == "" {
			r.EventRecorder.Event(cloudBucket, corev1.EventTypeNormal, "BucketExists", "Bucket already exists")
			log.Info("Bucket already exists", "bucketName", cloudBucket.Spec.BucketName)
		}
		cloudBucket.Status.BucketExists = true
		cloudBucket.Status.LastOperation = "Exists"
		cloudBucket.Status.ErrorMessage = ""
	}

	// Update status
	if err := r.Status().Update(ctx, cloudBucket); err != nil {
		log.Error(err, "Failed to update CloudBucket status")
		ErrorsTotal.Inc()
		r.EventRecorder.Event(cloudBucket, corev1.EventTypeWarning, "StatusUpdateFailed", fmt.Sprintf("Failed to update status: %v", err))
		return ctrl.Result{}, err
	}

	log.Info("Reconciliation completed", "bucketName", cloudBucket.Spec.BucketName, "status", cloudBucket.Status)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudBucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.EventRecorder = mgr.GetEventRecorderFor("cloud-storage-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1.CloudBucket{}).
		Complete(r)
}

// createBucket creates a new bucket in GCS
func (r *CloudBucketReconciler) createBucket(ctx context.Context, projectID, bucketName, location string) error {
	if bucketName == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	bucket := r.GCSClient.Bucket(bucketName)
	attrs := &storage.BucketAttrs{}
	if location != "" {
		attrs.Location = location
	}
	if err := bucket.Create(ctx, projectID, attrs); err != nil {
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	return nil
}

// deleteBucket deletes a bucket in GCS
func (r *CloudBucketReconciler) deleteBucket(ctx context.Context, bucketName string) error {
	if bucketName == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	bucket := r.GCSClient.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("Bucket(%q).Delete: %v", bucketName, err)
	}
	return nil
}

// bucketExists checks if a bucket exists in GCS
func (r *CloudBucketReconciler) bucketExists(ctx context.Context, bucketName string) (bool, error) {
	if bucketName == "" {
		return false, fmt.Errorf("bucket name cannot be empty")
	}
	bucket := r.GCSClient.Bucket(bucketName)
	_, err := bucket.Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist {
			return false, nil
		}
		return false, fmt.Errorf("Bucket(%q).Attrs: %v", bucketName, err)
	}
	return true, nil
}
