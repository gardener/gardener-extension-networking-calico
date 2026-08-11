// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ControllerConfiguration defines the configuration for the Calico networking extension.
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`
	// ClientConnection specifies the kubeconfig file and client connection
	// settings for the proxy server to use when communicating with the apiserver.
	// +optional
	ClientConnection *componentbaseconfigv1alpha1.ClientConnectionConfiguration `json:"clientConnection,omitempty"`
	// FeatureGates is a map of feature names to bools that enable
	// or disable alpha/experimental features.
	// Default: nil
	// +optional
	FeatureGates map[string]bool `json:"featureGates,omitempty"`
	// KubeAPIServerEndpoints contains the landscape-wide configuration for the kube-apiserver GlobalNetworkSet which
	// is deployed into shoot clusters.
	// +optional
	KubeAPIServerEndpoints *KubeAPIServerEndpointsConfiguration `json:"kubeAPIServerEndpoints,omitempty"`
}

// KubeAPIServerEndpointsConfiguration contains the landscape-wide configuration for the kube-apiserver
// GlobalNetworkSet.
type KubeAPIServerEndpointsConfiguration struct {
	// Enabled is the landscape-wide default which determines whether the GlobalNetworkSet is deployed into shoot
	// clusters. It can be overridden per shoot via the Network resource's providerConfig.
	// Default: false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}
