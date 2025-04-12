package controller

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// BucketsCreated counts the number of GCS buckets created
	BucketsCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cloud_storage_buckets_created_total",
			Help: "Total number of GCS buckets created",
		},
	)

	// BucketsRecreated counts the number of GCS buckets recreated
	BucketsRecreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cloud_storage_buckets_recreated_total",
			Help: "Total number of GCS buckets recreated after being missing",
		},
	)

	// BucketsDeleted counts the number of GCS buckets deleted
	BucketsDeleted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cloud_storage_buckets_deleted_total",
			Help: "Total number of GCS buckets deleted",
		},
	)

	// BucketsOrphaned counts the number of GCS buckets orphaned
	BucketsOrphaned = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cloud_storage_buckets_orphaned_total",
			Help: "Total number of GCS buckets orphaned",
		},
	)

	// ErrorsTotal counts the number of errors encountered
	ErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cloud_storage_errors_total",
			Help: "Total number of errors during bucket operations",
		},
	)
)

// init registers the metrics with the Prometheus registry
func init() {
	fmt.Println("Registering metrics")
	metrics.Registry.MustRegister(
		BucketsCreated,
		BucketsRecreated,
		BucketsDeleted,
		BucketsOrphaned,
		ErrorsTotal,
	)
}
