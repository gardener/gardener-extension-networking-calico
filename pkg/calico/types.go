// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package calico

import (
	"path/filepath"

	"github.com/gardener/gardener-extension-networking-calico/charts"
)

const (
	Name = "networking-calico"

	// AnnotationTyphaRestartedAt is set on the Network resource to trigger a calico-typha rolling
	// restart after a control plane migration or non-HA to HA transition. The value is an RFC3339
	// timestamp that is propagated into the Typha pod template annotation.
	AnnotationTyphaRestartedAt = "networking.calico.extensions.gardener.cloud/typha-migration-restart-at"

	// AnnotationControlPlaneHA records whether the shoot control plane had HighAvailability enabled
	// the last time the Network resource was reconciled. Used to detect non-HA to HA transitions.
	AnnotationControlPlaneHA = "networking.calico.extensions.gardener.cloud/control-plane-ha"

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
)

var (
	// CalicoChartPath path for internal Calico Chart
	CalicoChartPath = filepath.Join(charts.InternalChartsPath, "calico")

	// CalicoMonitoringChartPath  path for internal Calico monitoring chart
	CalicoMonitoringChartPath = filepath.Join(charts.InternalChartsPath, "calico-monitoring")
)
