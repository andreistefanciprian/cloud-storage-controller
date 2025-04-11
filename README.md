```

kubebuilder init --domain example.com --license apache2 --repo github.com/andreistefanciprian/cloud-storage-controller --project-name cloud-storage-controller --owner "Ciprian Andrei"

kubebuilder create api --group mygroup --version v1 --kind CloudBucket

make generate
make manifests

gcloud iam service-accounts create cloud-storage-controller \
    --project=$GCP_PROJECT \
    --display-name="Cloud Storage Controller"

gcloud projects add-iam-policy-binding $GCP_PROJECT \
    --member="serviceAccount:cloud-storage-controller@${GCP_PROJECT}.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

gcloud iam service-accounts add-iam-policy-binding \
    cloud-storage-controller@${GCP_PROJECT}.iam.gserviceaccount.com \
    --project=$GCP_PROJECT \
    --role="roles/iam.workloadIdentityUser" \
    --member="serviceAccount:${GCP_PROJECT}.svc.id.goog[default/cloud-storage-controller-controller-manager]"
```