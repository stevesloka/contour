// Copyright © 2018 Heptio
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

package k8s

import (
	"encoding/json"

	ingressroutev1 "github.com/heptio/contour/apis/contour/v1beta1"
	clientset "github.com/heptio/contour/internal/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// IngressRouteStatus allows for updating the object's Status field
type IngressRouteStatus struct {
	Client clientset.Interface
}

// SetValidStatus sets the IngressRoute status field to an Valid or Invalid status
func (irs *IngressRouteStatus) SetValidStatus(desc string, paths []string, existing *ingressroutev1.IngressRoute) error {
	updated := existing.DeepCopy()
	updated.Status = ingressroutev1.Status{
		CurrentStatus: "valid",
		Description:   desc,
	}

	return irs.setStatus(paths, existing, updated)
}

// SetInvalidStatus sets the IngressRoute status field to an Valid or Invalid status
func (irs *IngressRouteStatus) SetInvalidStatus(desc string, paths []string, existing *ingressroutev1.IngressRoute) error {
	updated := existing.DeepCopy()
	updated.Status = ingressroutev1.Status{
		CurrentStatus: "invalid",
		Description:   desc,
	}

	return irs.setStatus(paths, existing, updated)
}

// SetOrphanStatus sets the IngressRoute status field to an Orphan status
func (irs *IngressRouteStatus) SetOrphanStatus(desc string, existing *ingressroutev1.IngressRoute) error {
	updated := existing.DeepCopy()
	updated.Status = ingressroutev1.Status{
		CurrentStatus: "orphan",
		Description:   desc,
	}

	return irs.setStatus([]string{}, existing, updated)
}

func (irs *IngressRouteStatus) setStatus(paths []string, existing, updated *ingressroutev1.IngressRoute) error {
	existingBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	// Need to set the resource version of the updated endpoints to the resource
	// version of the current service. Otherwise, the resulting patch does not
	// have a resource version, and the server complains.
	updated.ResourceVersion = existing.ResourceVersion
	updatedBytes, err := json.Marshal(updated)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(existingBytes, updatedBytes, ingressroutev1.IngressRoute{})
	if err != nil {
		return err
	}

	_, err = irs.Client.ContourV1beta1().IngressRoutes(existing.GetNamespace()).Patch(existing.GetName(), types.MergePatchType, patchBytes)
	return err
}
