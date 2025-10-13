// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

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

const (
	IPAMCalico    string = "calico-ipam"
	IPAMHostLocal string = "host-local"
)

// IPv4 contains configuration for calico ipv4 specific settings
type IPv4 struct {
	// Pool configures the type of ip pool for the tunnel interface
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#environment-variables
	// +optional
	Pool *Pool `json:"pool,omitempty"`
	// Mode is the mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
	// ipip pools accept all pool mode values values
	// vxlan pools accept only Always and Never (unchecked)
	// +optional
	Mode *PoolMode `json:"mode,omitempty"`
	// AutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods
	// +optional
	AutoDetectionMethod *string `json:"autoDetectionMethod,omitempty"`
}

// IPv6 contains configuration for calico ipv6 specific settings
type IPv6 struct {
	// Pool configures the type of ip pool for the tunnel interface
	// https://docs.tigera.io/calico/latest/reference/configure-calico-node#configuring-the-default-ip-pools
	// +optional
	Pool *Pool `json:"pool,omitempty"`
	// Mode is the mode for the IPv6 Pool (e.g. Always, Never, CrossSubnet)
	// TODO: VXLAN also supports CrossSubnet for VXLAN. Why is this not supported?
	// vxlan pools accept only Always and Never (unchecked)
	// +optional
	Mode *PoolMode `json:"mode,omitempty"`
	// AutoDetectionMethod is the method to use to autodetect the IPv6 address for this host. This is only used when the IPv6 address is being autodetected.
	// https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods
	// +optional
	AutoDetectionMethod *string `json:"autoDetectionMethod,omitempty"`
	// SourceNATEnabled indicates whether the pod IP addresses should be masqueraded when targeting external destinations.
	// Per default, source network address translation is disabled.
	// +optional
	SourceNATEnabled *bool `json:"sourceNATEnabled,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkConfig configuration for the calico networking plugin
type NetworkConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Backend defines whether a backend should be used or not (e.g., bird or none)
	// +optional
	Backend *Backend `json:"backend,omitempty"`
	// IPAM to use for the Calico Plugin (e.g., host-local or calico-ipam)
	// +optional
	IPAM *IPAM `json:"ipam,omitempty"`
	// IPv4 contains configuration for calico ipv4 specific settings
	// +optional
	IPv4 *IPv4 `json:"ipv4,omitempty"`
	// IPv6 contains configuration for calico ipv4 specific settings
	// +optional
	IPv6 *IPv6 `json:"ipv6,omitempty"`
	// Typha settings to use for calico-typha component
	// +optional
	Typha *Typha `json:"typha,omitempty"`
	// VethMTU settings used to configure calico port mtu
	// +optional
	VethMTU *string `json:"vethMTU,omitempty"`
	// EbpfDataplane enables the eBPF dataplane mode.
	// +optional
	EbpfDataplane *EbpfDataplane `json:"ebpfDataplane,omitempty"`
	// Overlay enables the network overlay
	// +optional
	Overlay *Overlay `json:"overlay,omitempty"`
	// SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server (default: enabled)
	// +optional
	SnatToUpstreamDNS *SnatToUpstreamDNS `json:"snatToUpstreamDNS,omitempty"`
	// AutoScaling defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).
	// +optional
	AutoScaling *AutoScaling `json:"autoScaling,omitempty"`

	// VXLAN enables vxlan as overlay network
	VXLAN *VXLAN `json:"vxlan,omitempty"`

	// DEPRECATED.
	// IPIP is the IPIP Mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
	// It was moved into the IPv4 struct, kept for backwards compatibility.
	// Will be removed in a future Gardener release.
	// +optional
	IPIP *PoolMode `json:"ipip,omitempty"`
	// DEPRECATED.
	// IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
	// It was moved into the IPv4 struct, kept for backwards compatibility.
	// Will be removed in a future Gardener release.
	// +optional
	IPAutoDetectionMethod *string `json:"ipAutodetectionMethod,omitempty"`

	// WireguardEncryption is the option to enable node to node wireguard encryption
	WireguardEncryption bool `json:"wireguardEncryption,omitempty"`

	// BirdExporter configures the bird metrics exporter.
	BirdExporter *BirdExporter `json:"birdExporter,omitempty"`

	// Multus configures Multus CNI.
	Multus *Multus `json:"multus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkStatus contains information about created Network resources.
type NetworkStatus struct {
	metav1.TypeMeta `json:",inline"`
}

// IPAM defines the block that configuration for the ip assignment plugin to be used
type IPAM struct {
	// Type defines the IPAM plugin type
	Type string `json:"type"`
	// CIDR defines the CIDR block to be used
	// +optional
	CIDR *CIDR `json:"cidr,omitempty"`
}

// Typha defines the block with configurations for calico typha
type Typha struct {
	// Enabled is used to define whether calico-typha is required or not.
	// Note, typha is used to offload kubernetes API server,
	// thus consider not to disable it for large clusters in terms of node count.
	// More info can be found here https://docs.projectcalico.org/v3.9/reference/typha/
	Enabled bool `json:"enabled"`
}

type EbpfDataplane struct {
	// Enabled enables the eBPF dataplane mode.
	Enabled bool `json:"enabled"`
}

type Overlay struct {
	// Enabled enables the network overlay.
	Enabled bool `json:"enabled"`
	// CreatePodRoutes installs routes to pods on all cluster nodes.
	// This will only work if the cluster nodes share a single L2 network.
	// +optional
	CreatePodRoutes *bool `json:"createPodRoutes,omitempty"`
}

// SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server
type SnatToUpstreamDNS struct {
	Enabled bool `json:"enabled"`
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
	Mode AutoscalingMode `json:"mode"`
	// Resources optionally defines the amount of resources to statically allocate for the calico components in case of
	// static resource allocation.
	// In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate.
	// +optional
	Resources *Resources `json:"resources,omitempty"`
}

// Resources optionally defines the amount of resources to statically allocate for the calico components in case of
// static resource allocation.
// In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate.
type Resources struct {
	// Node optionally defines the amount of resources to statically allocate for the calico node component in case of
	// static resource allocation.
	// In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate for the calico
	// node component.
	// +optional
	Node *corev1.ResourceList `json:"node,omitempty"`
	// Node optionally defines the amount of resources to statically allocate for the calico typha component in case of
	// static resource allocation.
	// In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate for the calico
	// typha component.
	// +optional
	Typha *corev1.ResourceList `json:"typha,omitempty"`
}

type VXLAN struct {
	// Enabled enables vxlan as overlay network.
	Enabled bool `json:"enabled"`
}

type BirdExporter struct {
	// Enabled enables the bird metrics exporter.
	Enabled bool `json:"enabled"`
}

type Multus struct {
	// Enabled enables Multus CNI.
	Enabled bool `json:"enabled"`
	// InstallCNIPlugins enables the installation of containernetworking/plugins.
	// +optional
	InstallCNIPlugins *bool `json:"installCNIPlugins,omitempty"`
}
