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

	"github.com/projectcontour/contour/internal/annotation"
	"github.com/projectcontour/contour/internal/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

			if !b.delegationPermitted(secretName, ing.GetNamespace()) {
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

func (c *Contour) lookupSecureVirtualHost(name string) *SecureVirtualHost {
	svh, ok := c.securevirtualhosts[name]
	if !ok {
		svh := &SecureVirtualHost{
			VirtualHost: VirtualHost{
				Name: name,
			},
		}
		c.securevirtualhosts[svh.VirtualHost.Name] = svh
		return svh
	}
	return svh
}

// lookupSecret returns a Secret if present or nil if the underlying kubernetes
// secret fails validation or is missing.
func (b *Builder) lookupSecret(m types.NamespacedName, validate func(*v1.Secret) error) (*Secret, error) {
	sec, ok := b.Source.secrets[m]
	if !ok {
		return nil, fmt.Errorf("Secret not found")
	}

	if err := validate(sec); err != nil {
		return nil, err
	}

	s := &Secret{
		Object: sec,
	}

	b.secrets[k8s.NamespacedNameOf(sec)] = s
	return s, nil
}
