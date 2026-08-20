// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

// The CRD based API group is used because this extension does not deploy the Calico API server, hence
// `projectcalico.org/v3` does not exist in shoot clusters.
const (
	globalNetworkSetAPIVersion = "crd.projectcalico.org/v1"
	globalNetworkSetKind       = "GlobalNetworkSet"

	annotationSource = "networking.gardener.cloud/source"
)

type globalNetworkSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec globalNetworkSetSpec `json:"spec"`
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
		TypeMeta: metav1.TypeMeta{
			APIVersion: globalNetworkSetAPIVersion,
			Kind:       globalNetworkSetKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: calico.KubeAPIServerEndpointsName,
			// The set carries exactly one label: labels are selector input for Calico policies, so anything added here
			// would leak into policies which do not mean to refer to this set.
			Labels:      map[string]string{calico.KubeAPIServerEndpointsLabelKey: calico.KubeAPIServerEndpointsLabelValue},
			Annotations: map[string]string{annotationSource: string(endpoints.Source)},
		},
		Spec: globalNetworkSetSpec{Nets: endpoints.CIDRs},
	})
}
