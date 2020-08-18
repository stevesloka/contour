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

package dag

import (
	"fmt"
	"strings"

	"github.com/projectcontour/contour/internal/annotation"
	"github.com/projectcontour/contour/internal/k8s"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Contour struct {
	cache *KubernetesCache

	BuilderData
}

func (c *Contour) Build(cache *KubernetesCache) BuilderData {

	// setup secure vhosts if there is a matching secret
	// we do this first so that the set of active secure vhosts is stable
	// during computeIngresses.
	c.computeSecureVirtualhosts()

	c.computeIngresses()

	return BuilderData{}
}

// computeSecureVirtualhosts populates tls parameters of
// secure virtual hosts.
func (c *Contour) computeSecureVirtualhosts() {
	for _, ing := range c.cache.ingresses {
		for _, tls := range ing.Spec.TLS {
			secretName := k8s.NamespacedNameFrom(tls.SecretName, k8s.DefaultNamespace(ing.GetNamespace()))
			sec, err := c.cache.LookupSecret(secretName, validSecret)
			if err != nil {
				c.cache.WithField("name", ing.GetName()).
					WithField("namespace", ing.GetNamespace()).
					WithField("error", err.Error()).
					Errorf("invalid TLS secret %q", secretName)
				continue
			}

			if !c.cache.DelegationPermitted(secretName, ing.GetNamespace()) {
				c.cache.WithField("name", ing.GetName()).
					WithField("namespace", ing.GetNamespace()).
					WithField("error", err).
					Errorf("certificate delegation not permitted for Secret %q", secretName)
				continue
			}

			// We have validated the TLS secrets, so we can go
			// ahead and create the SecureVirtualHost for this
			// Ingress.
			for _, host := range tls.Hosts {
				svhost := c.lookupSecureVirtualHost(host)
				svhost.Secret = sec
				svhost.MinTLSVersion = annotation.MinTLSVersion(
					annotation.CompatAnnotation(ing, "tls-minimum-protocol-version"))
			}
		}
	}
}

func (c *Contour) computeIngresses() {
	// deconstruct each ingress into routes and virtualhost entries
	for _, ing := range c.cache.ingresses {

		// rewrite the default ingress to a stock ingress rule.
		rules := rulesFromSpec(ing.Spec)
		for _, rule := range rules {
			b.computeIngressRule(ing, rule)
		}
	}
}

func (c *Contour) computeIngressRule(ing *v1beta1.Ingress, rule v1beta1.IngressRule) {
	host := rule.Host
	if strings.Contains(host, "*") {
		// reject hosts with wildcard characters.
		return
	}
	if host == "" {
		// if host name is blank, rewrite to Envoy's * default host.
		host = "*"
	}
	for _, httppath := range httppaths(rule) {
		path := stringOrDefault(httppath.Path, "/")
		be := httppath.Backend
		m := types.NamespacedName{Name: be.ServiceName, Namespace: ing.Namespace}
		s, err := c.lookupService(m, be.ServicePort)
		if err != nil {
			continue
		}

		r := route(ing, path, s)

		// should we create port 80 routes for this ingress
		if annotation.TLSRequired(ing) || annotation.HTTPAllowed(ing) {
			c.lookupVirtualHost(host).addRoute(r)
		}

		// computeSecureVirtualhosts will have populated b.securevirtualhosts
		// with the names of tls enabled ingress objects. If host exists then
		// it is correctly configured for TLS.
		svh, ok := b.securevirtualhosts[host]
		if ok && host != "*" {
			svh.addRoute(r)
		}
	}
}

// lookupService returns a Service that matches the Meta and Port of the Kubernetes' Service,
// or an error if the service or port can't be located.
func (c *Contour) lookupService(m types.NamespacedName, port intstr.IntOrString) (*Service, error) {
	lookup := func() *Service {
		if port.Type != intstr.Int {
			// can't handle, give up
			return nil
		}
		sm := servicemeta{
			name:      m.Name,
			namespace: m.Namespace,
			port:      int32(port.IntValue()),
		}
		return c.services[sm]
	}

	s := lookup()
	if s != nil {
		return s, nil
	}
	svc, ok := c.cache.services[m]
	if !ok {
		return nil, fmt.Errorf("service %q not found", m)
	}
	for i := range svc.Spec.Ports {
		p := svc.Spec.Ports[i]
		switch {
		case int(p.Port) == port.IntValue():
			return b.addService(svc, p), nil
		case port.String() == p.Name:
			return b.addService(svc, p), nil
		}
	}
	return nil, fmt.Errorf("port %q on service %q not matched", port.String(), m)
}
