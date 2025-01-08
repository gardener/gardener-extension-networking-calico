// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package calico

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Backend string

const (
	Bird  Backend = "bird"
	None  Backend = "none"
	VXLan Backend = "vxlan"
)

type PoolMode string
type IPv4PoolMode = PoolMode

const (
	Always      PoolMode = "Always"
	Never       PoolMode = "Never"
	CrossSubnet PoolMode = "CrossSubnet"
	Off         PoolMode = "Off"
)

type CIDR string

type Pool string
type IPv4Pool = Pool

const (
	PoolIPIP  Pool = "ipip"
	PoolVXLan Pool = "vxlan"
)

// IPv4 contains configuration for calico ipv4 specific settings
type IPv4 struct {
	// Pool configures the type of ip pool for the tunnel interface.
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#environment-variables
	Pool *Pool
	// Mode is the mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
	// ipip pools accept all pool mode values values
	// vxlan pools accept only Always and Never (unchecked)
	Mode *PoolMode
	// AutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods
	AutoDetectionMethod *string
}

// IPv6 contains configuration for calico ipv6 specific settings
type IPv6 struct {
	// Pool configures the type of ip pool for the tunnel interface
	// https://docs.tigera.io/calico/latest/reference/configure-calico-node#configuring-the-default-ip-pools
	Pool *Pool
	// Mode is the mode for the IPv6 Pool (e.g. Always, Never, CrossSubnet)
	// vxlan pools accept only Always and Never (unchecked)
	Mode *PoolMode
	// AutoDetectionMethod is the method to use to autodetect the IPv6 address for this host. This is only used when the IPv6 address is being autodetected.
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods
	AutoDetectionMethod *string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkConfig configuration for the calico networking plugin
type NetworkConfig struct {
	metav1.TypeMeta
	// Backend defines whether a backend should be used or not (e.g., bird or none)
	Backend *Backend
	// IPAM to use for the Calico Plugin (e.g., host-local or Calico)
	IPAM *IPAM
	// IPv4 contains configuration for calico ipv4 specific settings
	IPv4 *IPv4
	// IPv6 contains configuration for calico ipv4 specific settings
	IPv6 *IPv6
	// Typha settings to use for calico-typha component
	Typha *Typha
	// VethMTU settings used to configure calico port mtu
	VethMTU *string
	// EbpfDataplane enables the eBPF dataplane mode.
	EbpfDataplane *EbpfDataplane
	// Overlay enables the network overlay
	Overlay *Overlay
	// SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server (default: enabled)
	SnatToUpstreamDNS *SnatToUpstreamDNS
	// AutoScaling defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).
	// +optional
	AutoScaling *AutoScaling

	// VXLAN enables vxlan as overlay network
	VXLAN *VXLAN

	// DEPRECATED.
	// IPIP is the IPIP Mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
	// It was moved into the IPv4 struct, kept for backwards compatibility.
	// Will be removed in a future Gardener release.
	IPIP *PoolMode
	// DEPRECATED.
	// IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
	// It was moved into the IPv4 struct, kept for backwards compatibility.
	// Will be removed in a future Gardener release.
	IPAutoDetectionMethod *string

	// WireguardEncryption is the option to enable node to node wireguard encryption
	WireguardEncryption bool
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkStatus contains information about created Network resources.
type NetworkStatus struct {
	metav1.TypeMeta
}

// IPAM defines the block that configuration for the ip assignment plugin to be used
type IPAM struct {
	// Type defines the IPAM plugin type
	Type string
	// CIDR defines the CIDR block to be used
	CIDR *CIDR
}

// Typha defines the block with configurations for calico typha
type Typha struct {
	// Enabled is used to define whether calico-typha is required or not.
	// Note, typha is used to offload kubernetes API server,
	// thus consider not to disable it for large clusters in terms of node count.
	// More info can be found here https://docs.projectcalico.org/v3.9/reference/typha/
	Enabled bool
}

type EbpfDataplane struct {
	// Enabled enables the eBPF dataplane mode.
	Enabled bool
}

type Overlay struct {
	// Enabled enables the network overlay.
	Enabled bool
	// CreatePodRoutes installs routes to pods on all cluster nodes.
	// This will only work if the cluster nodes share a single L2 network.
	CreatePodRoutes *bool
}

// SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server
type SnatToUpstreamDNS struct {
	Enabled bool
}

// AutoscalingMode is a type alias for the autoscaling mode string.
type AutoscalingMode string

const (
	// AutoscalingModeClusterProportional is a constant for cluster-proportional autoscaling mode.
	AutoscalingModeClusterProportional AutoscalingMode = "cluster-proportional"
	// AutoscalingModeVPA is a constant for vertical pod autoscaling mode.
	AutoscalingModeVPA AutoscalingMode = "vpa"
	// AutoscalingModeStatic is a constant for static resource allocation as autoscaling mode.
	AutoscalingModeStatic AutoscalingMode = "static"
)

// AutoScaling defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).
type AutoScaling struct {
	// Mode defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).
	Mode AutoscalingMode
	// Resources optionally defines the amount of resources to statically allocate for the calico components.
	Resources *StaticResources
}

// StaticResources optionally defines the amount of resources to statically allocate for the calico components.
type StaticResources struct {
	// Node optionally defines the amount of resources to statically allocate for the calico node component.
	Node *corev1.ResourceList
	// Node optionally defines the amount of resources to statically allocate for the calico typha component.
	Typha *corev1.ResourceList
}

type VXLAN struct {
	// Enabled enables vxlan as overlay network.
	Enabled bool
}
