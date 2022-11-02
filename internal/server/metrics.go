// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "bifrost"
)

type metrics struct {
	registry         *prometheus.Registry
	requestsDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
	m := &metrics{}
	m.registry = prometheus.NewRegistry()
	options := prometheus.ProcessCollectorOpts{
		Namespace: namespace,
	}
	m.registry.MustRegister(prometheus.NewProcessCollector(options))
	m.registry.MustRegister(prometheus.NewGoCollector())

	m.requestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "requests_duration",
			Help:      "Duration of the http requests.",
		},
		[]string{"method", "status_code", "installation_id"},
	)
	m.registry.MustRegister(m.requestsDuration)

	return m
}

func (m *metrics) observeRequest(method, installationID string, statusCode int, duration float64) {
	m.requestsDuration.With(
		prometheus.Labels{
			"method":          method,
			"status_code":     strconv.Itoa(statusCode),
			"installation_id": installationID,
		},
	).Observe(duration)
}

// metricsHandler returns the handler that is going to be used by the
// service to expose the metrics.
func (m *metrics) metricsHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Timeout:           30 * time.Second,
		EnableOpenMetrics: true,
	})
}
