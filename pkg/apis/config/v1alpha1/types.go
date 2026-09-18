// SPDX-FileCopyrightText: Contributors to the Gardener project
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
	// KubeAPIServerGlobalNetworkSet contains the operator configuration for the kube-apiserver GlobalNetworkSet which
	// is deployed into shoot clusters.
	// +optional
	KubeAPIServerGlobalNetworkSet *KubeAPIServerGlobalNetworkSetConfiguration `json:"kubeAPIServerGlobalNetworkSet,omitempty"`
}

// KubeAPIServerGlobalNetworkSetConfiguration contains the operator configuration for the kube-apiserver
// GlobalNetworkSet.
type KubeAPIServerGlobalNetworkSetConfiguration struct {
	// Enabled is the default for all shoots handled by this extension deployment. It determines whether the
	// GlobalNetworkSet is deployed into their clusters and can be overridden per shoot via the Network resource's
	// providerConfig.
	// Default: false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}
