package framework

import (
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// CreateDeployment creates a deployment resource in a namespace
func (f *Framework) CreateDeployment(deploy *appsv1.Deployment) error {
	_, err := f.K8sClientset.AppsV1().Deployments(deploy.GetNamespace()).Create(deploy)

	watch, err := f.K8sClientset.CoreV1().Pods(deploy.GetNamespace()).Watch(metav1.ListOptions{
		LabelSelector: labels.Set(deploy.Labels).String(),
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	go func() {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}
			fmt.Println(p.Status.ContainerStatuses)
			fmt.Println(p.Status.Phase)
		}
	}()
	time.Sleep(1 * time.Second)

	return err
}

// DeploymentManifest returns a deployment spec of a *appsv1.Deployment
func (f *Framework) DeploymentManifest(ns, name, image string) *appsv1.Deployment {

	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"app": "contour-integration-nginx",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "contour-integration-nginx",
				},
			},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"app": "contour-integration-nginx",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:            name,
						Image:           image,
						ImagePullPolicy: v1.PullIfNotPresent,
						Ports: []v1.ContainerPort{{
							Name:          "http",
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	}
}
