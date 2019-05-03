package integration

import (
	"fmt"

	"github.com/heptio/contour/integration/framework"
	"github.com/stretchr/testify/suite"
)

type IngressSuite struct {
	suite.Suite
	Framework        *framework.Framework
	Host             string
	EnvoyFQDN        string
	TestingNamespace string
}

// SetupSuite setup at the beginning of test
func (s *IngressSuite) SetupSuite() {
	// // Verify connectivity
	// err := framework.ValidateKubectl()
	// require.NoError(s.T(), err, "Could not connect to cluster")

	// // Deploy CRDs
	// err = framework.CreateCRDs()
	// require.NoError(s.T(), err, "Unable to deploy IngressRoute CRDs")

	// // Deploy Rbac
	// err = framework.DeployRbac()
	// require.NoError(s.T(), err, "Unable to deploy rbac permissions")

	// // Deploy Contour+Envoy
	// err = framework.DeployContourEnvoy()
	// require.NoError(s.T(), err, "Unable to deploy Contour+Envoy")
}

// TearDownSuite teardown at the end of test
func (s *IngressSuite) TearDownSuite() {
	fmt.Println("---- Tear down suite!!")
}
