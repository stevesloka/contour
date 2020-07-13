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

package contour

import (
	"sort"
	"strings"
	"sync"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/projectcontour/contour/internal/envoy"
	"github.com/projectcontour/contour/internal/protobuf"
	"github.com/projectcontour/contour/internal/sorter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scache "k8s.io/client-go/tools/cache"
)

//// A EndpointsTranslator translates Kubernetes Endpoints objects into Envoy
//// ClusterLoadAssignment objects.
//type EndpointsTranslator struct {
//	logrus.FieldLogger
//	ClusterLoadAssignmentCache
//
//}

func (ch *CacheHandler) OnAdd(obj interface{}) {
	switch obj := obj.(type) {
	case *v1.Endpoints:
		ch.addEndpoints(obj)
	default:
		ch.Errorf("OnAddEndpoint unexpected type %T: %#v", obj, obj)
	}
}

func (ch *CacheHandler) OnUpdateEndpoint(oldObj, newObj interface{}) {
	switch newObj := newObj.(type) {
	case *v1.Endpoints:
		oldObj, ok := oldObj.(*v1.Endpoints)
		if !ok {
			ch.Errorf("OnUpdateEndpoints %#v received invalid oldObj %T; %#v", newObj, oldObj, oldObj)
			return
		}
		ch.updateEndpoints(oldObj, newObj)
	default:
		e.Errorf("OnUpdate unexpected type %T: %#v", newObj, newObj)
	}
}

func (ch *CacheHandler) OnDelete(obj interface{}) {
	switch obj := obj.(type) {
	case *v1.Endpoints:
		ch.removeEndpoints(obj)
	case k8scache.DeletedFinalStateUnknown:
		ch.OnDelete(obj.Obj) // recurse into ourselves with the tombstoned value
	default:
		ch.Errorf("OnDelete unexpected type %T: %#v", obj, obj)
	}
}

func (ch *CacheHandler) Contents() []proto.Message {
	values := ch.ClusterLoadAssignmentCache.Contents()
	sort.Stable(sorter.For(values))
	return protobuf.AsMessages(values)
}

func (ch *CacheHandler) Query(names []string) []proto.Message {
	e.clusterLoadAssignmentCache.mu.Lock()
	defer e.clusterLoadAssignmentCache.mu.Unlock()
	values := make([]*v2.ClusterLoadAssignment, 0, len(names))
	for _, n := range names {
		v, ok := e.entries[n]
		if !ok {
			v = &v2.ClusterLoadAssignment{
				ClusterName: n,
			}
		}
		values = append(values, v)
	}

	sort.Stable(sorter.For(values))
	return protobuf.AsMessages(values)
}

func (*CacheHandler) TypeURL() string { return resource.EndpointType }

func (ch *CacheHandler) addEndpoints(ep *v1.Endpoints) {
	if ch.ClusterCache.Contains(ep.Name) {
		ch.recomputeClusterLoadAssignment(nil, ep)
	}
}

func (ch *CacheHandler) updateEndpoints(oldep, newep *v1.Endpoints) {
	if ch.ClusterCache.Contains(newep.Name) {
		if len(newep.Subsets) == 0 && len(oldep.Subsets) == 0 {
			// if there are no endpoints in this object, and the old
			// object also had zero endpoints, ignore this update
			// to avoid sending a noop notification to watchers.
			return
		}
		ch.recomputeClusterLoadAssignment(oldep, newep)
	}
}

func (ch *CacheHandler) removeEndpoints(ep *v1.Endpoints) {
	ch.recomputeClusterLoadAssignment(ep, nil)
}

// recomputeClusterLoadAssignment recomputes the EDS cache taking into account old and new endpoints.
func (ch *CacheHandler) recomputeClusterLoadAssignment(oldep, newep *v1.Endpoints) {
	// skip computation if either old and new services or endpoints are equal (thus also handling nil)
	if oldep == newep {
		return
	}

	if oldep == nil {
		oldep = &v1.Endpoints{
			ObjectMeta: newep.ObjectMeta,
		}
	}

	if newep == nil {
		newep = &v1.Endpoints{
			ObjectMeta: oldep.ObjectMeta,
		}
	}

	seen := make(map[string]bool)
	// add or update endpoints
	for _, s := range newep.Subsets {
		if len(s.Addresses) < 1 {
			// skip subset without ready addresses.
			continue
		}
		for _, p := range s.Ports {
			if p.Protocol != "TCP" {
				// skip non TCP ports
				continue
			}

			addresses := append([]v1.EndpointAddress{}, s.Addresses...) // shallow copy
			sort.Slice(addresses, func(i, j int) bool { return addresses[i].IP < addresses[j].IP })

			lbendpoints := make([]*envoy_api_v2_endpoint.LbEndpoint, 0, len(addresses))
			for _, a := range addresses {
				addr := envoy.SocketAddress(a.IP, int(p.Port))
				lbendpoints = append(lbendpoints, envoy.LBEndpoint(addr))
			}

			cla := &v2.ClusterLoadAssignment{
				ClusterName: servicename(newep.ObjectMeta, p.Name),
				Endpoints: []*envoy_api_v2_endpoint.LocalityLbEndpoints{{
					LbEndpoints: lbendpoints,
				}},
			}
			seen[cla.ClusterName] = true
			ch.Add(cla)
		}
	}

	// iterate over the ports in the old spec, remove any were not seen.
	for _, s := range oldep.Subsets {
		if len(s.Addresses) == 0 {
			continue
		}
		for _, p := range s.Ports {
			name := servicename(oldep.ObjectMeta, p.Name)
			if _, ok := seen[name]; !ok {
				// port is no longer present, remove it.
				e.Remove(name)
			}
		}
	}

}

type ClusterLoadAssignmentCache struct {
	mu      sync.Mutex
	entries map[string]*v2.ClusterLoadAssignment
	Cond
}

// Add adds an entry to the cache. If a ClusterLoadAssignment with the same
// name exists, it is replaced.
func (c *ClusterLoadAssignmentCache) Add(a *v2.ClusterLoadAssignment) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = make(map[string]*v2.ClusterLoadAssignment)
	}
	c.entries[a.ClusterName] = a
	c.Notify(a.ClusterName)
}

// Remove removes the named entry from the cache. If the entry
// is not present in the cache, the operation is a no-op.
func (c *ClusterLoadAssignmentCache) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, name)
	c.Notify(name)
}

// Contents returns a copy of the contents of the cache.
func (c *ClusterLoadAssignmentCache) Contents() []*v2.ClusterLoadAssignment {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]*v2.ClusterLoadAssignment, 0, len(c.entries))
	for _, v := range c.entries {
		values = append(values, v)
	}
	return values
}

//// Contains returns `true` if the item exists in the cache
//func (c *ClusterLoadAssignmentCache) Contains(name string) bool {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//	if c.entries == nil {
//		return false
//	}
//	if _, ok := c.entries[name]; !ok {
//		return false
//	}
//	return true
//}

// servicename returns the name of the cluster this meta and port
// refers to. The CDS name of the cluster may include additional suffixes
// but these are not known to EDS.
func servicename(meta metav1.ObjectMeta, portname string) string {
	name := []string{
		meta.Namespace,
		meta.Name,
		portname,
	}
	if portname == "" {
		name = name[:2]
	}
	return strings.Join(name, "/")
}
