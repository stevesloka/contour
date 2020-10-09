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

package v3

import (
	envoy_api_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_v3_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/projectcontour/contour/internal/protobuf"
)

// UpstreamTLSTransportSocket returns a custom transport socket using the UpstreamTlsContext provided.
func UpstreamTLSTransportSocket(tls *envoy_v3_tls.UpstreamTlsContext) *envoy_api_v3_core.TransportSocket {
	return &envoy_api_v3_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_api_v3_core.TransportSocket_TypedConfig{
			TypedConfig: protobuf.MustMarshalAny(tls),
		},
	}
}

// DownstreamTLSTransportSocket returns a custom transport socket using the DownstreamTlsContext provided.
func DownstreamTLSTransportSocket(tls *envoy_v3_tls.DownstreamTlsContext) *envoy_api_v3_core.TransportSocket {
	return &envoy_api_v3_core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &envoy_api_v3_core.TransportSocket_TypedConfig{
			TypedConfig: protobuf.MustMarshalAny(tls),
		},
	}
}
