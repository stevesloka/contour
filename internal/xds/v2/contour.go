// Copyright Project Contour Authors
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

package v2

import (
	"context"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/projectcontour/contour/internal/xds"
	"github.com/sirupsen/logrus"
)

type grpcStream interface {
	Context() context.Context
	Send(*envoy_api_v2.DiscoveryResponse) error
	Recv() (*envoy_api_v2.DiscoveryRequest, error)
}

// NewContourServer creates an internally implemented Server that streams the
// provided set of Resource objects. The returned Server implements the xDS
// State of the World (SotW) variant.
func NewContourServer(log logrus.FieldLogger, notifier xds.Notifier, resources ...xds.Resource) Server {
	c := contourServer{
		FieldLogger: log,
		resources:   map[string]xds.Resource{},
		Notifier:    notifier,
	}

	for i, r := range resources {
		c.resources[r.TypeURL()] = resources[i]
	}

	return &c
}

type contourServer struct {
	// Since we only implement the streaming state of the world
	// protocol, embed the default null implementations to handle
	// the unimplemented gRPC endpoints.
	discovery.UnimplementedAggregatedDiscoveryServiceServer
	discovery.UnimplementedSecretDiscoveryServiceServer
	envoy_api_v2.UnimplementedRouteDiscoveryServiceServer
	envoy_api_v2.UnimplementedEndpointDiscoveryServiceServer
	envoy_api_v2.UnimplementedClusterDiscoveryServiceServer
	envoy_api_v2.UnimplementedListenerDiscoveryServiceServer

	logrus.FieldLogger
	resources map[string]xds.Resource
	xds.Notifier
}

//type eventNotifier struct {
//	logrus.FieldLogger
//	event chan xds.EnvoyMessage
//}

// stream processes a stream of DiscoveryRequests.
func (s *contourServer) stream(st grpcStream) error {

	// Notify whether the stream terminated on error.
	done := func(log *logrus.Entry, err error) error {
		if err != nil {
			log.WithError(err).Error("stream terminated")
		} else {
			log.Info("stream terminated")
		}

		return err
	}

	// now stick in this loop until the client disconnects.
	for {
		// first we wait for the request from Envoy, this is part of
		// the xDS protocol.
		req, err := st.Recv()
		if err != nil {
			return done(s.WithField("connection", ""), err)
		}
		s.Event <- xds.EnvoyMessage{
			TypeUrl: req.TypeUrl,
			//ErrorDetail:   req.ErrorDetail,
			ResponseNonce: req.ResponseNonce,
			VersionInfo:   req.VersionInfo,
			ResourceNames: req.ResourceNames,
		}
	}
}

func (s *contourServer) StreamClusters(srv envoy_api_v2.ClusterDiscoveryService_StreamClustersServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamEndpoints(srv envoy_api_v2.EndpointDiscoveryService_StreamEndpointsServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamListeners(srv envoy_api_v2.ListenerDiscoveryService_StreamListenersServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamRoutes(srv envoy_api_v2.RouteDiscoveryService_StreamRoutesServer) error {
	return s.stream(srv)
}

func (s *contourServer) StreamSecrets(srv discovery.SecretDiscoveryService_StreamSecretsServer) error {
	return s.stream(srv)
}
