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
	"github.com/projectcontour/contour/internal/annotation"
	"github.com/projectcontour/contour/internal/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

type Ingress struct {
	Source *KubernetesCache
	logrus.FieldLogger

	services map[servicemeta]*Service
	secrets  map[types.NamespacedName]*Secret

	virtualhosts       map[string]*VirtualHost
	securevirtualhosts map[string]*SecureVirtualHost

	orphaned map[types.NamespacedName]bool
}

func NewIngress(cache *KubernetesCache) *Ingress {
	return &Ingress{
		Source:      cache,
		FieldLogger: logrus.New(),
	}
}

func (i *Ingress) Build(dag *DAG) *DAG {

	// setup secure vhosts if there is a matching secret
	// we do this first so that the set of active secure vhosts is stable
	// during computeIngresses.
	i.computeSecureVirtualhosts()

	//
	//c.computeIngresses()

	return dag
}

// computeSecureVirtualhosts populates tls parameters of
// secure virtual hosts.
func (i *Ingress) computeSecureVirtualhosts() {
	for _, ing := range i.Source.ingresses {
		for _, tls := range ing.Spec.TLS {
			secretName := k8s.NamespacedNameFrom(tls.SecretName, k8s.DefaultNamespace(ing.GetNamespace()))
			sec, err := i.Source.LookupSecret(secretName, validSecret)
			if err != nil {
				i.WithError(err).
					WithField("name", ing.GetName()).
					WithField("namespace", ing.GetNamespace()).
					WithField("secret", secretName).
					Error("unresolved secret reference")
				continue
			}
			i.secrets[k8s.NamespacedNameOf(sec.Object)] = sec

			if !i.Source.DelegationPermitted(secretName, ing.GetNamespace()) {
				i.WithError(err).
					WithField("name", ing.GetName()).
					WithField("namespace", ing.GetNamespace()).
					WithField("secret", secretName).
					Error("certificate delegation not permitted")
				continue
			}

			// We have validated the TLS secrets, so we can go
			// ahead and create the SecureVirtualHost for this
			// Ingress.
			for _, host := range tls.Hosts {
				svhost := b.lookupSecureVirtualHost(host)
				svhost.Secret = sec
				svhost.MinTLSVersion = annotation.MinTLSVersion(
					annotation.CompatAnnotation(ing, "tls-minimum-protocol-version"))
			}
		}
	}
}
