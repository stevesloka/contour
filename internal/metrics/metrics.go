// Copyright © 2017 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ContourMetrics provide Prometheus metrics for contour
type ContourMetrics struct {
	metrics map[string]prometheus.Collector
	log     logrus.FieldLogger
}

const (
	// IngressRouteUpstreamEndpointsTotal represents a Prometheus metric which reports
	// the total number of endpoints for an IngressRoute
	IngressRouteUpstreamEndpointsTotal = "contour_ingressroute_upstream_endpoints_total"

	// IngressRouteUpstreamEndpointsHealthy represents a Prometheus metric which reports
	// the total number of healthy endpoints for an IngressRoute
	IngressRouteUpstreamEndpointsHealthy = "contour_ingressroute_upstream_endpoints_healthy"

	// IngressRouteUpstreamEndpointsChange represents a Prometheus metric which reports
	// the number of endpoint changes for an IngressRoute
	IngressRouteUpstreamEndpointsChange = "contour_ingressroute_upstream_endpoints_change"
)

// NewMetrics returns a map of Prometheus metrics
func NewMetrics(log logrus.FieldLogger) ContourMetrics {
	return ContourMetrics{
		log: log,
		metrics: map[string]prometheus.Collector{
			IngressRouteUpstreamEndpointsTotal: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: IngressRouteUpstreamEndpointsTotal,
					Help: "Total number of endpoints for an IngressRoute",
				},
				[]string{"namespace", "name"},
			),
			IngressRouteUpstreamEndpointsHealthy: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: IngressRouteUpstreamEndpointsHealthy,
					Help: "Total number of healthy endpoints for an IngressRoute",
				},
				[]string{"namespace", "name"},
			),
			IngressRouteUpstreamEndpointsChange: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: IngressRouteUpstreamEndpointsChange,
					Help: "Number of endpoint changes for an IngressRoute",
				},
				[]string{"namespace", "name"},
			),
		},
	}
}

// RegisterPrometheus registers the metrics
func (c *ContourMetrics) RegisterPrometheus() {
	// Register with Prometheus's default registry
	for _, v := range c.metrics {
		prometheus.MustRegister(v)
	}
}

// IngressRouteUpstreamEndpointsTotalMetric formats a total endpoint prometheus metric
func (c *ContourMetrics) IngressRouteUpstreamEndpointsTotalMetric(namespace, name string, totalEndpoints int) {
	m, ok := c.metrics[IngressRouteUpstreamEndpointsTotal].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, name).Set(float64(totalEndpoints))
	} else {
		c.log.Errorln("Could not get ref to 'IngressRouteUpstreamEndpointsTotal' metric")
	}
}

// IngressRouteUpstreamEndpointsHealthyMetric formats a healthy endpoint prometheus metric
func (c *ContourMetrics) IngressRouteUpstreamEndpointsHealthyMetric(namespace, name string, totalEndpoints int) {
	m, ok := c.metrics[IngressRouteUpstreamEndpointsHealthy].(*prometheus.GaugeVec)
	if ok {
		m.WithLabelValues(namespace, name).Set(float64(totalEndpoints))
	} else {
		c.log.Errorln("Could not get ref to 'IngressRouteUpstreamEndpointsHealthy' metric")
	}

}

// IngressRouteUpstreamEndpointsChangeMetric formats an endpoint counter prometheus metric
func (c *ContourMetrics) IngressRouteUpstreamEndpointsChangeMetric(namespace, name string) {
	m, ok := c.metrics[IngressRouteUpstreamEndpointsChange].(*prometheus.CounterVec)
	if ok {
		m.WithLabelValues(namespace, name).Inc()
	} else {
		c.log.Errorln("Could not get ref to 'IngressRouteUpstreamEndpointsChange' metric")
	}
}
