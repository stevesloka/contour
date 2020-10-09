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
	envoy_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

// SocketAddress creates a new TCP envoy_v3_core.Address.
func SocketAddress(address string, port int) *envoy_v3_core.Address {
	if address == "::" {
		return &envoy_v3_core.Address{
			Address: &envoy_v3_core.Address_SocketAddress{
				SocketAddress: &envoy_v3_core.SocketAddress{
					Protocol:   envoy_v3_core.SocketAddress_TCP,
					Address:    address,
					Ipv4Compat: true,
					PortSpecifier: &envoy_v3_core.SocketAddress_PortValue{
						PortValue: uint32(port),
					},
				},
			},
		}
	}
	return &envoy_v3_core.Address{
		Address: &envoy_v3_core.Address_SocketAddress{
			SocketAddress: &envoy_v3_core.SocketAddress{
				Protocol: envoy_v3_core.SocketAddress_TCP,
				Address:  address,
				PortSpecifier: &envoy_v3_core.SocketAddress_PortValue{
					PortValue: uint32(port),
				},
			},
		},
	}
}
