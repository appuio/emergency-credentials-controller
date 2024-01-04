package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const MetricsNamespace = "emergency_credentials_controller"

var (
	verifiedTokensValidUntil = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Name:      "verified_tokens_valid_until_seconds",
			Help:      "The latest valid_until timestamp for verified tokens for the emergency account.",
		},
		[]string{"emergency_account"},
	)
)

func deleteVerifiedTokensValidUntil(emergencyAccount string) {
	verifiedTokensValidUntil.Delete(prometheus.Labels{"emergency_account": emergencyAccount})
}

func init() {
	metrics.Registry.MustRegister(verifiedTokensValidUntil)
}
