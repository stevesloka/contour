package framework

import (
	ingressv1beta1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress creates an ingress resource in a namespace
func (f *Framework) CreateIngress(ing *ingressv1beta1.Ingress) error {
	_, err := f.K8sClientset.ExtensionsV1beta1().Ingresses(ing.GetNamespace()).Create(ing)
	return err
}

// DeleteAllIngress cleans up all Ingress resources for a namespace
func (f *Framework) DeleteAllIngress(namespace string) error {
	return f.K8sClientset.ExtensionsV1beta1().Ingresses(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{})
}
