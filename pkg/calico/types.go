// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package calico

import (
	"path/filepath"

	"github.com/gardener/gardener-extension-networking-calico/charts"
)

const (
	Name = "networking-calico"

	// ImageNames
	CNIImageName                                   = "calico-cni"
	NodeImageName                                  = "calico-node"
	KubeControllersImageName                       = "calico-kube-controllers"
	TyphaImageName                                 = "calico-typha"
	CalicoClusterProportionalAutoscalerImageName   = "calico-cpa"
	ClusterProportionalVerticalAutoscalerImageName = "calico-cpva"
	BirdExporterImageName                          = "bird-exporter"
	MultusImageName                                = "multus-cni"
	CNIPluginsImageName                            = "cni-plugins"

	// MonitoringChartName
	MonitoringName = "calico-monitoring-config"

	// ReleaseName is the name of the Calico Release
	ReleaseName = "calico"

	// The following constants describe the GlobalNetworkSet rendered by the apiserverendpoints package. They live here
	// so that the NetworkConfig validation can refer to the reserved label key without importing that package.

	// KubeAPIServerEndpointsDefaultName is the default name of the GlobalNetworkSet which contains the IP addresses
	// of the shoot's kube-apiserver endpoint.
	KubeAPIServerEndpointsDefaultName = "gardener-kube-apiserver"
	// KubeAPIServerEndpointsLabelKey is the key of the label which is always set on the kube-apiserver
	// GlobalNetworkSet. It forms the contract for (Global)NetworkPolicies referencing the set and must not be
	// overridden by user provided labels.
	KubeAPIServerEndpointsLabelKey = "networking.gardener.cloud/endpoint"
	// KubeAPIServerEndpointsLabelValue is the value of KubeAPIServerEndpointsLabelKey.
	KubeAPIServerEndpointsLabelValue = "kube-apiserver"
)

var (
	// CalicoChartPath path for internal Calico Chart
	CalicoChartPath = filepath.Join(charts.InternalChartsPath, "calico")

	// CalicoMonitoringChartPath  path for internal Calico monitoring chart
	CalicoMonitoringChartPath = filepath.Join(charts.InternalChartsPath, "calico-monitoring")
)
