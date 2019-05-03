// +build integration

package integration

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/heptio/contour/integration/framework"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	ingressv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var inCluster bool
var kubeconfig string
var host string
var envoyFQDN string

func init() {
	inCluster = getenv_bool("INCLUSTER", true)
	kubeconfig = getenv_string("KUBECONFIG", "")
	host = getenv_string("HOST", "containersteve.com")
	envoyFQDN = getenv_string("ENVOYFQDN", "http://envoy")
}

type ingressSuiteTest struct {
	IngressSuite
}

func TestIngressSuite(t *testing.T) {
	client, contourClient, err := framework.NewClient(kubeconfig, inCluster)
	if err != nil {
		log.Fatalf("Could not get clients: %v", err)
	}

	ingressSuite := &ingressSuiteTest{
		IngressSuite{
			Host:      host,
			EnvoyFQDN: envoyFQDN,
			Framework: &framework.Framework{
				K8sClientset:     client,
				ContourClientset: contourClient,
			},
		},
	}
	suite.Run(t, ingressSuite)
}

func (s *ingressSuiteTest) SetupTest() {
	log.Println("Starting a Test.")
	err := s.Framework.DeleteAllIngress("heptio-contour")
	require.NoError(s.T(), err)
	log.Println("Test Init Complete")
}

func (s *ingressSuiteTest) TearDownTest() {
	// log.Println("Starting a Test. Cleaning up existing objects")
	// log.Println("Objects Cleaned up Successfully")
}

func (m *ingressSuiteTest) TestSingleService() {
	type testCase struct {
		Name           string
		Ingress        *ingressv1beta1.Ingress
		ExpectedResult int
	}

	// Create test service
	err := m.Framework.CreateService(m.Framework.ServiceManifest("heptio-contour", "nginx", v1.ServicePort{
		Protocol:   "TCP",
		Port:       80,
		TargetPort: intstr.FromInt(80),
	}))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// Create test deployment
	err = m.Framework.CreateDeployment(m.Framework.DeploymentManifest("heptio-contour", "nginx", "nginx"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	arrTestcase := []testCase{
		{
			Name: "simple",
			Ingress: &ingressv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "heptio-contour",
					Name:      "nginx",
				},
				Spec: ingressv1beta1.IngressSpec{
					Backend: &ingressv1beta1.IngressBackend{
						ServiceName: "nginx",
						ServicePort: intstr.FromInt(80),
					},
				},
			},
			ExpectedResult: 200,
		},
	}

	for _, tc := range arrTestcase {
		m.T().Run(tc.Name, func(t *testing.T) {
			// Create Ingress resource
			m.Framework.CreateIngress(tc.Ingress)

			req, err := http.NewRequest("GET", m.IngressSuite.EnvoyFQDN, nil)
			require.NoError(t, err)

			req.Host = m.IngressSuite.Host

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedResult, resp.StatusCode, tc.Name)
			defer resp.Body.Close()
		})
	}
}

func getenv_string(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getenv_bool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		val, _ := strconv.ParseBool(value)
		return val
	}
	return fallback
}
