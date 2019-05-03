package framework

import (
	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteAllIngressRoutes cleans up all IngressRoute resources for a namespace
func (f *Framework) DeleteAllIngressRoutes(namespace string) error {
	return f.ContourClientset.ContourV1beta1().IngressRoutes(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{})
}

// CreateIngressRoute creates an ingress resource in a namespace
func (f *Framework) CreateIngressRoute(ing *ingressroutev1.IngressRoute) error {
	_, err := f.ContourClientset.ContourV1beta1().IngressRoutes(ing.GetNamespace()).Create(ing)
	return err
}
