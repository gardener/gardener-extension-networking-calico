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

	// MonitoringChartName
	MonitoringName = "calico-monitoring-config"

	// ReleaseName is the name of the Calico Release
	ReleaseName = "calico"
)

var (
	// CalicoChartPath path for internal Calico Chart
	CalicoChartPath = filepath.Join(charts.InternalChartsPath, "calico")

	// CalicoMonitoringChartPath  path for internal Calico monitoring chart
	CalicoMonitoringChartPath = filepath.Join(charts.InternalChartsPath, "calico-monitoring")
)
