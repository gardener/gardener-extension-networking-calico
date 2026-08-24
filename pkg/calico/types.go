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
	// restart after a control plane migration. The RFC3339 timestamp value is propagated into the
	// Typha pod template annotation to cause a rolling restart.
	AnnotationTyphaRestartedAt = "networking.calico.extensions.gardener.cloud/typha-restarted-at"

	// AnnotationTyphaRestartReason is set alongside AnnotationTyphaRestartedAt to explain why
	// the restart was triggered. Visible on both the Network resource and the Typha pod template.
	AnnotationTyphaRestartReason = "networking.calico.extensions.gardener.cloud/typha-restart-reason"

	// TyphaRestartReasonHATransition is the restart reason for a non-HA to HA transition.
	TyphaRestartReasonHATransition = "ha-transition"

	// AnnotationCalicoNodeRestartedAt is set on the Network resource to trigger a calico-node rolling
	// restart after a control plane migration when Typha is disabled. The RFC3339 timestamp value
	// is propagated into the calico-node DaemonSet pod template annotation to cause a rolling restart.
	AnnotationCalicoNodeRestartedAt = "networking.calico.extensions.gardener.cloud/calico-node-restarted-at"

	// AnnotationCalicoNodeRestartReason is set alongside AnnotationCalicoNodeRestartedAt to explain why
	// the restart was triggered.
	AnnotationCalicoNodeRestartReason = "networking.calico.extensions.gardener.cloud/calico-node-restart-reason"

	// RestartReasonCPM is the restart reason recorded on both Typha and calico-node after a
	// control plane migration.
	RestartReasonCPM = "control-plane-migration"

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
