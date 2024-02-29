// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	"encoding/json"
	"fmt"
	"strconv"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/gardener/gardener-extension-networking-calico/imagevector"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

const (
	hostLocal    = "host-local"
	calicoIPAM   = "calico-ipam"
	usePodCIDR   = "usePodCidr"
	usePodCIDRv6 = "usePodCidrIPv6"
	defaultMTU   = "1440"
)

type calicoConfig struct {
	Backend         calicov1alpha1.Backend `json:"backend"`
	Felix           felix                  `json:"felix"`
	IPv4            ipv4                   `json:"ipv4"`
	IPv6            ipv6                   `json:"ipv6"`
	IPAM            ipam                   `json:"ipam"`
	Typha           typha                  `json:"typha"`
	KubeControllers kubeControllers        `json:"kubeControllers"`
	VethMTU         string                 `json:"veth_mtu"`
	Monitoring      monitoring             `json:"monitoring"`
	NonPrivileged   bool                   `json:"nonPrivileged"`
}

type felix struct {
	IPInIP                      felixIPinIP                      `json:"ipinip"`
	BPF                         felixBPF                         `json:"bpf"`
	BPFKubeProxyIptablesCleanup felixBPFKubeProxyIptablesCleanup `json:"bpfKubeProxyIPTablesCleanup"`
}

type felixIPinIP struct {
	Enabled bool `json:"enabled"`
}

type felixBPF struct {
	Enabled bool `json:"enabled"`
}

type felixBPFKubeProxyIptablesCleanup struct {
	Enabled bool `json:"enabled"`
}

type ipv4 struct {
	Enabled             bool                    `json:"enabled"`
	Pool                calicov1alpha1.Pool     `json:"pool"`
	Mode                calicov1alpha1.PoolMode `json:"mode"`
	AutoDetectionMethod *string                 `json:"autoDetectionMethod"`
}

type ipv6 struct {
	Enabled             bool                    `json:"enabled"`
	Pool                calicov1alpha1.Pool     `json:"pool"`
	Mode                calicov1alpha1.PoolMode `json:"mode"`
	AutoDetectionMethod *string                 `json:"autoDetectionMethod"`
	NATOutgoing         bool                    `json:"natOutgoing"`
}

type ipam struct {
	IPAMType   string `json:"type"`
	Subnet     string `json:"subnet"`
	AssignIPv4 bool   `json:"assign_ipv4"`
	AssignIPv6 bool   `json:"assign_ipv6"`
}

type kubeControllers struct {
	Enabled bool `json:"enabled"`
}

type monitoring struct {
	Enabled bool `json:"enabled"`
	// TyphaPort is the port used to expose typha metrics
	TyphaMetricsPort string `json:"typhaMetricsPort"`
	// FelixPort is the port used to exposed felix metrics
	FelixMetricsPort string `json:"felixMetricsPort"`
}

type typha struct {
	Enabled bool `json:"enabled"`
}

var defaultCalicoConfig = calicoConfig{
	Backend: calicov1alpha1.Bird,
	Felix: felix{
		IPInIP: felixIPinIP{
			Enabled: true,
		},
		BPF: felixBPF{
			Enabled: false,
		},
		BPFKubeProxyIptablesCleanup: felixBPFKubeProxyIptablesCleanup{
			Enabled: false,
		},
	},
	IPAM: ipam{
		IPAMType:   hostLocal,
		AssignIPv4: false,
		AssignIPv6: false,
	},
	Typha: typha{
		Enabled: true,
	},
	KubeControllers: kubeControllers{
		Enabled: true,
	},
	VethMTU: defaultMTU,
	Monitoring: monitoring{
		Enabled:          true,
		FelixMetricsPort: "9091",
		TyphaMetricsPort: "9093",
	},
}

func newCalicoConfig() calicoConfig {
	return defaultCalicoConfig
}

func (c *calicoConfig) toMap() (map[string]interface{}, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not marshal calico config: %v", err)
	}
	var configMap map[string]interface{}
	err = json.Unmarshal(bytes, &configMap)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal calico config: %v", err)
	}
	return configMap, nil
}

// ComputeCalicoChartValues computes the values for the calico chart.
func ComputeCalicoChartValues(
	network *extensionsv1alpha1.Network,
	config *calicov1alpha1.NetworkConfig,
	kubernetesVersion string,
	wantsVPA bool,
	kubeProxyEnabled bool,
	isPSPDisabled bool,
	nonPrivileged bool,
	nodeCIDR *string,
) (map[string]interface{}, error) {
	typedConfig, err := generateChartValues(network, config, kubeProxyEnabled, nonPrivileged)
	if err != nil {
		return nil, fmt.Errorf("error when generating calico config: %v", err)
	}
	calicoConfig, err := typedConfig.toMap()
	if err != nil {
		return nil, fmt.Errorf("could not convert calico config: %v", err)
	}
	calicoChartValues := map[string]interface{}{
		"vpa": map[string]interface{}{
			"enabled": wantsVPA,
		},
		"images": map[string]interface{}{
			calico.CNIImageName:                                   imagevector.CalicoCNIImage(kubernetesVersion),
			calico.TyphaImageName:                                 imagevector.CalicoTyphaImage(kubernetesVersion),
			calico.KubeControllersImageName:                       imagevector.CalicoKubeControllersImage(kubernetesVersion),
			calico.NodeImageName:                                  imagevector.CalicoNodeImage(kubernetesVersion),
			calico.CalicoClusterProportionalAutoscalerImageName:   imagevector.ClusterProportionalAutoscalerImage(kubernetesVersion),
			calico.ClusterProportionalVerticalAutoscalerImageName: imagevector.ClusterProportionalVerticalAutoscalerImage(kubernetesVersion),
		},
		"global": map[string]string{
			"podCIDR": network.Spec.PodCIDR,
		},
		"config":      calicoConfig,
		"pspDisabled": isPSPDisabled,
	}

	if nodeCIDR != nil {
		calicoChartValues["global"].(map[string]string)["nodeCIDR"] = *nodeCIDR
	}

	if config != nil && config.Overlay != nil {
		calicoChartValues["global"].(map[string]string)["overlayEnabled"] = strconv.FormatBool(config.Overlay.Enabled)
	}

	if config != nil && config.Overlay != nil && !config.Overlay.Enabled {
		// Overlay is disabled => enable source NAT to upstream DNS per default
		snatToUpstreamDNS := true
		if config.SnatToUpstreamDNS != nil {
			snatToUpstreamDNS = config.SnatToUpstreamDNS.Enabled
		}
		calicoChartValues["global"].(map[string]string)["snatToUpstreamDNSEnabled"] = strconv.FormatBool(snatToUpstreamDNS)
	}

	if config != nil && config.AutoScaling != nil && config.AutoScaling.Mode == calicov1alpha1.AutoscalingModeVPA && wantsVPA {
		calicoChartValues["vpa"].(map[string]interface{})["node"] = strconv.FormatBool(true)
		calicoChartValues["vpa"].(map[string]interface{})["typha"] = strconv.FormatBool(true)
	}
	return calicoChartValues, nil
}

func generateChartValues(network *extensionsv1alpha1.Network, config *calicov1alpha1.NetworkConfig, kubeProxyEnabled bool, nonPrivileged bool) (*calicoConfig, error) {
	// by default assume IPv4 (dual-stack is not supported, yet)
	ipFamilies := sets.New[extensionsv1alpha1.IPFamily](network.Spec.IPFamilies...)
	isIPv4 := true
	isIPv6 := false
	if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv6) {
		isIPv4 = false
		isIPv6 = true
	}

	c := newCalicoConfig()
	if isIPv4 {
		c.IPAM.AssignIPv4 = true
		c.IPAM.Subnet = usePodCIDR
		c.IPv4 = ipv4{
			Enabled:             true,
			Pool:                calicov1alpha1.PoolIPIP,
			Mode:                calicov1alpha1.Always,
			AutoDetectionMethod: nil,
		}
	}

	if isIPv6 {
		c.IPAM.AssignIPv6 = true
		c.IPAM.Subnet = usePodCIDRv6
		// NOTE: calico guide (https://docs.tigera.io/calico/latest/networking/ipam/ipv6#enable-ipv6-only) requires ipam
		// type to be calico-ipam for single IPv6 or dual-stack clusters.
		c.IPAM.IPAMType = calicoIPAM
		c.IPv6 = ipv6{
			Enabled:             true,
			Pool:                calicov1alpha1.PoolVXLan,
			Mode:                calicov1alpha1.Never,
			AutoDetectionMethod: nil,
			NATOutgoing:         true,
		}
		c.Felix.IPInIP.Enabled = false
	}

	if !kubeProxyEnabled {
		c.Felix.BPFKubeProxyIptablesCleanup.Enabled = true
	}

	// will be overridden to false if config.EbpfDataplane.Enabled==true
	c.NonPrivileged = nonPrivileged

	return mergeCalicoValuesWithConfig(&c, config, isIPv4, isIPv6)
}

func mergeCalicoValuesWithConfig(c *calicoConfig, config *calicov1alpha1.NetworkConfig, isIPv4, isIPv6 bool) (*calicoConfig, error) {
	if config == nil {
		return c, nil
	}

	if config.Backend != nil {
		switch *config.Backend {
		case calicov1alpha1.Bird, calicov1alpha1.VXLan, calicov1alpha1.None:
			c.Backend = *config.Backend
		default:
			return nil, fmt.Errorf("unsupported value for backend: %s", *config.Backend)
		}
	}
	if c.Backend == calicov1alpha1.None {
		c.KubeControllers.Enabled = false
		c.Felix.IPInIP.Enabled = false
		c.IPv4.Mode = calicov1alpha1.Never
	}

	if config.EbpfDataplane != nil && config.EbpfDataplane.Enabled {
		c.Felix.BPF.Enabled = true
		c.NonPrivileged = false
	}

	if config.IPAM != nil && config.IPAM.Type != "" {
		c.IPAM.IPAMType = config.IPAM.Type
	}

	if c.IPAM.IPAMType == hostLocal {
		if config.IPAM != nil && config.IPAM.CIDR != nil {
			c.IPAM.Subnet = string(*config.IPAM.CIDR)
		}
	}

	if config.IPv4 != nil {
		if !isIPv4 {
			return nil, fmt.Errorf("IPv4 configuration must not be specified if Shoot doesn't use IPv4 networking")
		}

		if config.IPv4.Pool != nil {
			switch *config.IPv4.Pool {
			case calicov1alpha1.PoolIPIP, calicov1alpha1.PoolVXLan:
				c.IPv4.Pool = *config.IPv4.Pool
			default:
				return nil, fmt.Errorf("unsupported value for ipv4 pool: %s", *config.IPv4.Pool)
			}
		}
		if config.IPv4.Mode != nil {
			switch *config.IPv4.Mode {
			case calicov1alpha1.Always, calicov1alpha1.Never, calicov1alpha1.Off, calicov1alpha1.CrossSubnet:
				c.IPv4.Mode = *config.IPv4.Mode
			default:
				return nil, fmt.Errorf("unsupported value for ipv4 mode: %s", *config.IPv4.Mode)
			}
		}
		if config.IPv4.AutoDetectionMethod != nil {
			c.IPv4.AutoDetectionMethod = config.IPv4.AutoDetectionMethod
		}
	} else {
		// fallback to deprecated configuration fields
		// will be removed in a future Gardener release
		if config.IPIP != nil {
			if !isIPv4 {
				return nil, fmt.Errorf("IPv4 configuration must not be specified if Shoot doesn't use IPv4 networking")
			}
			switch *config.IPIP {
			case calicov1alpha1.Always, calicov1alpha1.Never, calicov1alpha1.Off, calicov1alpha1.CrossSubnet:
				c.IPv4.Mode = *config.IPIP
			default:
				return nil, fmt.Errorf("unsupported value for ipip: %s", *config.IPIP)
			}
		}
		if config.IPAutoDetectionMethod != nil {
			c.IPv4.AutoDetectionMethod = config.IPAutoDetectionMethod
		}
	}

	if config.IPv6 != nil {
		if !isIPv6 {
			return nil, fmt.Errorf("IPv6 configuration must not be specified if Shoot doesn't use IPv6 networking")
		}

		if config.IPv6.Pool != nil {
			switch *config.IPv6.Pool {
			case calicov1alpha1.PoolVXLan:
				c.IPv6.Pool = *config.IPv6.Pool
			default:
				return nil, fmt.Errorf("unsupported value for ipv6 pool: %s", *config.IPv6.Pool)
			}
		}
		if config.IPv6.Mode != nil {
			switch *config.IPv6.Mode {
			case calicov1alpha1.Always, calicov1alpha1.Never, calicov1alpha1.CrossSubnet:
				c.IPv6.Mode = *config.IPv6.Mode
			default:
				return nil, fmt.Errorf("unsupported value for ipv6 mode: %s", *config.IPv6.Mode)
			}
		}
		if config.IPv6.AutoDetectionMethod != nil {
			c.IPv6.AutoDetectionMethod = config.IPv6.AutoDetectionMethod
		}
	}

	if config.Typha != nil {
		c.Typha.Enabled = config.Typha.Enabled
	}

	if config.VethMTU != nil {
		c.VethMTU = *config.VethMTU
	}

	return c, nil
}
