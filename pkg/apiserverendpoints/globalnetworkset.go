// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"errors"
	"maps"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
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

// Name returns the name of the GlobalNetworkSet.
func Name(config *calicov1alpha1.KubeAPIServerEndpoints) string {
	if config != nil && config.Name != nil && *config.Name != "" {
		return *config.Name
	}

	return calico.KubeAPIServerEndpointsDefaultName
}

// RenderGlobalNetworkSet renders the GlobalNetworkSet holding the given endpoints as YAML. The result is a pure function
// of its input, in particular it carries no timestamp: the bytes are hashed into an immutable ManagedResource secret, so
// anything changing per reconciliation would re-apply the object in every shoot cluster.
func RenderGlobalNetworkSet(config *calicov1alpha1.KubeAPIServerEndpoints, endpoints *Endpoints) ([]byte, error) {
	if endpoints == nil || len(endpoints.CIDRs) == 0 {
		return nil, errors.New("refusing to render a GlobalNetworkSet without CIDRs")
	}

	return yaml.Marshal(&globalNetworkSet{
		APIVersion: globalNetworkSetAPIVersion,
		Kind:       globalNetworkSetKind,
		Metadata: globalNetworkSetMeta{
			Name:   Name(config),
			Labels: labels(config),
			Annotations: map[string]string{
				annotationPorts:  joinPorts(endpoints.Ports),
				annotationSource: string(endpoints.Source),
			},
		},
		Spec: globalNetworkSetSpec{Nets: endpoints.CIDRs},
	})
}

// labels merges the configured labels with the contract label, which cannot be overridden. Nothing else is added: the
// labels of a GlobalNetworkSet are selector input for Calico policies, so any extra label leaks into policies which do
// not mean to refer to this set.
func labels(config *calicov1alpha1.KubeAPIServerEndpoints) map[string]string {
	result := map[string]string{}

	if config != nil {
		maps.Copy(result, config.Labels)
	}

	result[calico.KubeAPIServerEndpointsLabelKey] = calico.KubeAPIServerEndpointsLabelValue

	return result
}

func joinPorts(ports []int32) string {
	result := make([]string, 0, len(ports))
	for _, port := range ports {
		result = append(result, strconv.FormatInt(int64(port), 10))
	}

	return strings.Join(result, ",")
}
