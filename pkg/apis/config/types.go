// SPDX-FileCopyrightText: Contributors to the Gardener project
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
	// KubeAPIServerGlobalNetworkSet contains the operator configuration for the kube-apiserver GlobalNetworkSet which
	// is deployed into shoot clusters.
	KubeAPIServerGlobalNetworkSet *KubeAPIServerGlobalNetworkSetConfiguration
}

// KubeAPIServerGlobalNetworkSetConfiguration contains the operator configuration for the kube-apiserver
// GlobalNetworkSet.
type KubeAPIServerGlobalNetworkSetConfiguration struct {
	// Enabled is the default for all shoots handled by this extension deployment. It determines whether the
	// GlobalNetworkSet is deployed into their clusters and can be overridden per shoot via the Network resource's
	// providerConfig.
	Enabled *bool
}
