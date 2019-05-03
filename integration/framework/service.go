package framework

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService creates an service resource in a namespace
func (f *Framework) CreateService(svc *v1.Service) error {
	_, err := f.K8sClientset.CoreV1().Services(svc.GetNamespace()).Create(svc)
	return err
}

// ServiceManifest returns a service spec of a *v1.Service
func (f *Framework) ServiceManifest(ns, name string, ports ...v1.ServicePort) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"app": "contour-integration-nginx",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "contour-integration-nginx",
			},
			Ports: ports,
		},
	}
}
