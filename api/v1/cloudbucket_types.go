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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudBucketSpec defines the desired state of CloudBucket
type CloudBucketSpec struct {
	// ProjectID is the GCP project ID where the bucket will be created.
	//+kubebuilder:validation:Required
	ProjectID string `json:"projectID"`

	// DeletePolicy determines whether the bucket is deleted when the CloudBucket resource is deleted.
	// Valid values are "Delete" (delete the bucket) or "Orphan" (leave the bucket).
	// If not specified, defaults to "Orphan".
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Enum=Delete;Orphan
	//+kubebuilder:default=Orphan
	DeletePolicy string `json:"deletePolicy,omitempty"`

	// Location is the GCS region or multi-region where the bucket is stored (e.g., "us", "eu", "asia")
	//+kubebuilder:validation:Optional
	Location string `json:"location,omitempty"`
}

// CloudBucketStatus defines the observed state of CloudBucket
type CloudBucketStatus struct {
	// BucketExists indicates whether the bucket exists in GCP.
	BucketExists bool `json:"bucketExists"`

	// BucketName is the actual name of the bucket created in GCP.
	//+kubebuilder:validation:Optional
	BucketName string `json:"bucketName,omitempty"`

	// LastOperation describes the last action performed by the controller (e.g., "Created", "Deleted", "Failed").
	//+kubebuilder:validation:Optional
	LastOperation string `json:"lastOperation,omitempty"`

	// ErrorMessage contains details of any error encountered during reconciliation.
	//+kubebuilder:validation:Optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CloudBucket is the Schema for the cloudbuckets API
type CloudBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudBucketSpec   `json:"spec,omitempty"`
	Status CloudBucketStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CloudBucketList contains a list of CloudBucket
type CloudBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudBucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudBucket{}, &CloudBucketList{})
}
