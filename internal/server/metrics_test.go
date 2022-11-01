// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package server

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	prometheusModels "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	metrics := newMetrics()
	server := New(Config{
		ServiceSettings: ServiceSettings{
			Host:        "localhost:12345",
			ServiceHost: "localhost:12346",
		},
	})

	go func() {
		if err := server.Start(); err != nil {
			require.NoError(t, err)
		}
	}()
	defer server.Stop()

	t.Run("Should store metrics for requests duration", func(t *testing.T) {
		m := &prometheusModels.Metric{}
		data, err := metrics.requestsDuration.GetMetricWith(
			prometheus.Labels{
				"method":          "method",
				"status_code":     "200",
				"installation_id": "",
			})
		require.NoError(t, err)
		require.NoError(t, data.(prometheus.Histogram).Write(m))
		require.Equal(t, uint64(0), m.Histogram.GetSampleCount())
		require.Equal(t, 0.0, m.Histogram.GetSampleSum())

		metrics.observeRequest("GET", "random_id", 200, 1.0)
		data, err = metrics.requestsDuration.GetMetricWith(
			prometheus.Labels{
				"method":          "GET",
				"installation_id": "random_id",
				"status_code":     "200",
			})
		require.NoError(t, err)
		require.NoError(t, data.(prometheus.Histogram).Write(m))
		require.Equal(t, uint64(1), m.Histogram.GetSampleCount())
		require.InDelta(t, 1, m.Histogram.GetSampleSum(), 0.001)
	})
}
