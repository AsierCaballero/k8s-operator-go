package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type OperatorMetrics struct {
	ReconcileDuration       *prometheus.HistogramVec
	ReconcileErrors         *prometheus.CounterVec
	ResourceOperationsTotal *prometheus.CounterVec
	AppDeploymentPhase      *prometheus.GaugeVec
}

var reconcileDurationBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}

func NewOperatorMetrics() *OperatorMetrics {
	m := &OperatorMetrics{}

	m.ReconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "k8s_operator_reconcile_duration_seconds",
		Help:    "Duration of reconciliation cycles",
		Buckets: reconcileDurationBuckets,
	}, []string{"result"})

	m.ReconcileErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_operator_reconcile_errors_total",
		Help: "Total number of reconciliation errors",
	}, []string{"reason"})

	m.ResourceOperationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_operator_resource_operations_total",
		Help: "Total number of Kubernetes resource operations",
	}, []string{"resource", "operation"})

	m.AppDeploymentPhase = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "k8s_operator_appdeployment_phase",
		Help: "Current phase of AppDeployment resources",
	}, []string{"name", "namespace", "phase"})

	metrics.Registry.MustRegister(
		m.ReconcileDuration,
		m.ReconcileErrors,
		m.ResourceOperationsTotal,
		m.AppDeploymentPhase,
	)

	return m
}

func (m *OperatorMetrics) ObserveReconcile(result string, duration float64) {
	m.ReconcileDuration.WithLabelValues(result).Observe(duration)
}

func (m *OperatorMetrics) RecordReconcileError(reason string) {
	m.ReconcileErrors.WithLabelValues(reason).Inc()
}

func (m *OperatorMetrics) RecordResourceOperation(resource, operation string) {
	m.ResourceOperationsTotal.WithLabelValues(resource, operation).Inc()
}

func (m *OperatorMetrics) SetDeploymentPhase(name, namespace, phase string) {
	m.AppDeploymentPhase.WithLabelValues(name, namespace, phase).Set(1)
}

func (m *OperatorMetrics) ResetDeploymentPhases(name, namespace string) {
	m.AppDeploymentPhase.DeletePartialMatch(prometheus.Labels{
		"name": name, "namespace": namespace,
	})
}
