// Copyright © 2020 VMware
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

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"

	"github.com/sirupsen/logrus"

	"github.com/projectcontour/contour/internal/workgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

func (s *shutdownmanagerContext) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *shutdownmanagerContext) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	prometheusURL := fmt.Sprintf("http://%s:%d%s", s.envoyHost, s.envoyPort, s.prometheusPath)
	envoyAdminURL := fmt.Sprintf("http://%s:%d/healthcheck/fail", s.envoyHost, s.envoyPort)

	// Send shutdown signal to Envoy to start draining connections
	err := shutdownEnvoy(envoyAdminURL)
	if err != nil {
		s.Error(err)
	}

	for {
		openConnections, err := getOpenConnections(prometheusURL, s.prometheusStat)
		if err != nil {
			s.Error(err)
		} else {
			fmt.Println("-- open connections: ", openConnections)
			if openConnections <= s.minOpenConnections {
				s.Infof("Found [%d] open connections with min number of [%d]", openConnections, s.minOpenConnections)
				return
			}
		}
		time.Sleep(s.checkInterval)
	}
}

func shutdownEnvoy(url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating GET request for URL %q failed: %s", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing GET request for URL %q failed: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET request for URL %q returned HTTP status %s", url, resp.Status)
	}
	return nil
}

func getOpenConnections(url, prometheusStat string) (int, error) {
	openConnections := 0
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return -1, fmt.Errorf("creating GET request for URL %q failed: %s", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("executing GET request for URL %q failed: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("GET request for URL %q returned HTTP status %s", url, resp.Status)
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("reading text format failed: %v", err)
	}
	for _, mf := range metricFamilies {
		if *mf.Name == prometheusStat {
			for _, metrics := range mf.Metric {
				for _, labels := range metrics.Label {
					switch *labels.Value {
					case "ingress_http", "ingress_https":
						openConnections += int(*metrics.Gauge.Value)
					}
				}
			}
		}
	}
	return openConnections, nil
}

//sum(envoy_http_downstream_cx_active{envoy_http_conn_manager_prefix="ingress_http"}) by (kubernetes_pod_name)

func doShutdownManager(config *shutdownmanagerContext) error {
	var g workgroup.Group

	g.Add(func(stop <-chan struct{}) error {
		config.Info("started envoy shutdown manager")
		defer config.Info("stopped")

		http.HandleFunc("/healthz", config.healthzHandler)
		http.HandleFunc("/shutdown", config.shutdownHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.httpServePort), nil))

		return nil
	})

	return g.Run()
}

// registerShutdownManager registers the envoy shutdown sub-command and flags
func registerShutdownManager(cmd *kingpin.CmdClause, log logrus.FieldLogger) (*kingpin.CmdClause, *shutdownmanagerContext) {
	ctx := &shutdownmanagerContext{
		FieldLogger: log,
	}
	shutdownmgr := cmd.Command("shutdown-manager", "Start envoy shutdown-manager.")
	shutdownmgr.Flag("check-interval", "Time to poll Envoy for open connections.").Default("2s").DurationVar(&ctx.checkInterval)
	shutdownmgr.Flag("min-open-connections", "Min number of open connections when polling Envoy.").Default("0").IntVar(&ctx.minOpenConnections)
	shutdownmgr.Flag("serve-port", "Port to serve the http server on.").Default("8090").IntVar(&ctx.httpServePort)
	shutdownmgr.Flag("prometheus-path", "The path to query Envoy's Prometheus HTTP Endpoint.").Default("/stats/prometheus").StringVar(&ctx.prometheusPath)
	shutdownmgr.Flag("prometheus-stat", "Prometheus stat to look query.").Default("envoy_http_downstream_cx_active").StringVar(&ctx.prometheusStat)
	shutdownmgr.Flag("envoy-host", "HTTP endpoint for Envoy's stats page.").Default("localhost").StringVar(&ctx.envoyHost)
	shutdownmgr.Flag("envoy-port", "HTTP port for Envoy's stats page.").Default("9001").IntVar(&ctx.envoyPort)

	return shutdownmgr, ctx
}
