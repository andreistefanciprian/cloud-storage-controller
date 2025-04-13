
### README.md

# Cloud Storage Controller

A Kubernetes controller to manage Google Cloud Storage (GCS) buckets with a `CloudBucket` custom resource.
Runs on GKE with Workload Identity for GCS access.

```
apiVersion: mygroup.example.com/v1
kind: CloudBucket
metadata:
  name: my-bucket-1
spec:
  projectID: rich-mountain-428806-r0
  deletePolicy: Delete
  location: asia
```

## What It Does
- Creates GCS buckets based on `CloudBucket` specs.
- Recreates buckets if deleted outside Kubernetes.
- Deletes buckets or leaves them based on `deletePolicy` (`Delete` or `Orphan`).

## Quick Start

```
# Create a Google Cloud Service Account (GSA) named "cloud-storage-controller" in your GCP project
gcloud iam service-accounts create cloud-storage-controller \
    --project=$GCP_PROJECT \
    --display-name="Cloud Storage Controller"

# Grant the GSA the "storage.admin" role to manage GCS buckets in the project
gcloud projects add-iam-policy-binding $GCP_PROJECT \
    --member="serviceAccount:cloud-storage-controller@${GCP_PROJECT}.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# Allow the KSA "controller-manager" in the "cloud-storage-controller-system" namespace
# to impersonate the GSA (alternative namespace binding, if used)
gcloud iam service-accounts add-iam-policy-binding \
    cloud-storage-controller@${GCP_PROJECT}.iam.gserviceaccount.com \
    --project=$GCP_PROJECT \
    --role="roles/iam.workloadIdentityUser" \
    --member="serviceAccount:${GCP_PROJECT}.svc.id.goog[cloud-storage-controller-system/controller-manager]"

# Create a temporary JSON key for the GSA for local testing (e.g., with "make run")
gcloud iam service-accounts keys create temp-sa-key.json \
    --iam-account=cloud-storage-controller@${GCP_PROJECT}.iam.gserviceaccount.com \
    --project=$GCP_PROJECT

# Set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the key file path
# for local authentication with the GCS client
export GOOGLE_APPLICATION_CREDENTIALS=$(pwd)/temp-sa-key.json

# test from local laptop
make manifests                                               
kubectl apply -f config/crd/bases/mygroup.example.com_cloudbuckets.yaml
make build
make run
k apply -f config/samples/mygroup_v1_cloudbucket.yaml
k delete -f config/samples/mygroup_v1_cloudbucket.yaml

# test in the cluster
make deploy
k logs -l control-plane=controller-manager -f -n cloud-storage-controller-system
k apply -f config/samples/mygroup_v1_cloudbucket.yaml
k delete -f config/samples/mygroup_v1_cloudbucket.yaml

# Check prometheus metrics
controller=`k get pods -n cloud-storage-controller-system --no-headers -l control-plane=controller-manager | awk '{print $1}'`
k port-forward pod/$controller 8080:8080
http://localhost:8080/metrics
```

## Other commands

```
kubebuilder init --domain example.com --license apache2 --repo github.com/andreistefanciprian/cloud-storage-controller --project-name cloud-storage-controller --owner "Ciprian Andrei"

kubebuilder create api --group mygroup --version v1 --kind CloudBucket

make generate
make manifests
```