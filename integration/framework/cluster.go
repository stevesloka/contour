package framework

import (
	"os/exec"
)

// ValidateKubectl validates a cluster
func ValidateKubectl() error {
	cmd := exec.Command("kubectl", "cluster-info")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	print(string(stdout))
	return nil
}

// CreateCRDs applies CRDs to cluster
func CreateCRDs() error {
	cmd := exec.Command("kubectl", "apply", "-f", "../../deployment/ds-hostnet-split/01-common.yaml")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	print(string(stdout))
	return nil
}

// DeployRbac from customized build
func DeployRbac() error {
	cmd := exec.Command("kubectl", "apply", "-f", "../../deployment/ds-hostnet-split/02-rbac.yaml")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	print(string(stdout))
	return nil
}

// DeployContourEnvoy from customized build
func DeployContourEnvoy() error {
	cmd := exec.Command("kubectl", "apply", "-f", "../integrationtest.yaml")
	//cmd := exec.Command("kubectl", "kustomize", "../../deployment/ds-hostnet-split/", "|", "kubectl", "apply", "-f", "-")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	print(string(stdout))
	return nil
}
