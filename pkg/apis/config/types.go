// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfig "k8s.io/component-base/config/v1alpha1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ControllerConfiguration defines the configuration for the Calico networking extension.
type ControllerConfiguration struct {
	metav1.TypeMeta
	// ClientConnection specifies the kubeconfig file and client connection
	// settings for the proxy server to use when communicating with the apiserver.
	ClientConnection *componentbaseconfig.ClientConnectionConfiguration
	// FeatureGates is a map of feature names to bools that enable
	// or disable alpha/experimental features.
	// Default: nil
	FeatureGates map[string]bool
	// KubeAPIServerEndpoints contains the landscape-wide configuration for the kube-apiserver GlobalNetworkSet which
	// is deployed into shoot clusters.
	KubeAPIServerEndpoints *KubeAPIServerEndpointsConfiguration
}

// KubeAPIServerEndpointsConfiguration contains the landscape-wide configuration for the kube-apiserver
// GlobalNetworkSet.
type KubeAPIServerEndpointsConfiguration struct {
	// Enabled is the landscape-wide default which determines whether the GlobalNetworkSet is deployed into shoot
	// clusters. It can be overridden per shoot via the Network resource's providerConfig.
	Enabled *bool
}
