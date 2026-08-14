// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"errors"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

// The CRD based API group is used because this extension does not deploy the Calico API server, hence
// `projectcalico.org/v3` does not exist in shoot clusters.
const (
	globalNetworkSetAPIVersion = "crd.projectcalico.org/v1"
	globalNetworkSetKind       = "GlobalNetworkSet"

	annotationPorts  = "networking.gardener.cloud/ports"
	annotationSource = "networking.gardener.cloud/source"
)

type globalNetworkSet struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Metadata   globalNetworkSetMeta `json:"metadata"`
	Spec       globalNetworkSetSpec `json:"spec"`
}

type globalNetworkSetMeta struct {
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type globalNetworkSetSpec struct {
	Nets []string `json:"nets"`
}

// RenderGlobalNetworkSet renders the GlobalNetworkSet holding the given endpoints as YAML. The result is a pure function
// of its input, in particular it carries no timestamp: the bytes are hashed into an immutable ManagedResource secret, so
// anything changing per reconciliation would re-apply the object in every shoot cluster.
func RenderGlobalNetworkSet(endpoints *Endpoints) ([]byte, error) {
	if endpoints == nil || len(endpoints.CIDRs) == 0 {
		return nil, errors.New("refusing to render a GlobalNetworkSet without CIDRs")
	}

	return yaml.Marshal(&globalNetworkSet{
		APIVersion: globalNetworkSetAPIVersion,
		Kind:       globalNetworkSetKind,
		Metadata: globalNetworkSetMeta{
			Name: calico.KubeAPIServerEndpointsName,
			// The set carries exactly one label: labels are selector input for Calico policies, so anything added here
			// would leak into policies which do not mean to refer to this set.
			Labels: map[string]string{calico.KubeAPIServerEndpointsLabelKey: calico.KubeAPIServerEndpointsLabelValue},
			Annotations: map[string]string{
				annotationPorts:  joinPorts(endpoints.Ports),
				annotationSource: string(endpoints.Source),
			},
		},
		Spec: globalNetworkSetSpec{Nets: endpoints.CIDRs},
	})
}

func joinPorts(ports []int32) string {
	result := make([]string, 0, len(ports))
	for _, port := range ports {
		result = append(result, strconv.FormatInt(int64(port), 10))
	}

	return strings.Join(result, ",")
}
