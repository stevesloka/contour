package framework

import (
	"fmt"
	"os"

	clientset "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Framework automates actions around the e2e tests
type Framework struct {
	K8sClientset     *kubernetes.Clientset
	ContourClientset *clientset.Clientset
}

// NewClient gens a new k8s client & a contour client
func NewClient(kubeconfig string, inCluster bool) (*kubernetes.Clientset, *clientset.Clientset, error) {
	var err error
	var config *rest.Config
	if kubeconfig != "" && !inCluster {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		Check(err)
	} else {
		config, err = rest.InClusterConfig()
		Check(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	contourClient, err := clientset.NewForConfig(config)
	return client, contourClient, nil
}

func Check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
