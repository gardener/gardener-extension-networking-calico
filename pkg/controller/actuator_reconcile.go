// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"net"
	"strings"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-networking-calico/charts"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	calicov1alpha1helper "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1/helper"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	chartspkg "github.com/gardener/gardener-extension-networking-calico/pkg/charts"
	"github.com/gardener/gardener-extension-networking-calico/pkg/features"
)

const (
	// CalicoConfigManagedResourceName is the name of the managed resource of networking calico
	CalicoConfigManagedResourceName = "extension-networking-calico-config"
)

func applyMonitoringConfig(ctx context.Context, seedClient client.Client, chartApplier gardenerkubernetes.ChartApplier, network *extensionsv1alpha1.Network, deleteChart bool) error {
	calicoControlPlaneMonitoringChart := &chart.Chart{
		Name:       calico.MonitoringName,
		EmbeddedFS: charts.InternalChart,
		Path:       calico.CalicoMonitoringChartPath,
		Objects: []*chart.Object{
			{
				Type: &corev1.ConfigMap{},
				Name: calico.MonitoringName,
			},
			{
				Type: &corev1.ConfigMap{},
				Name: "calico-dashboards",
			},
			{
				Type: &monitoringv1alpha1.ScrapeConfig{},
				Name: "shoot-calico-felix",
			},
			{
				Type: &monitoringv1alpha1.ScrapeConfig{},
				Name: "shoot-calico-typha",
			},
		},
	}

	if deleteChart {
		return client.IgnoreNotFound(calicoControlPlaneMonitoringChart.Delete(ctx, seedClient, network.Namespace))
	}

	// TODO(rfranzke): Delete this after August 2024.
	gep19Monitoring := seedClient.Get(ctx, client.ObjectKey{Name: "prometheus-shoot", Namespace: network.Namespace}, &appsv1.StatefulSet{}) == nil
	if gep19Monitoring {
		if err := kubernetesutils.DeleteObject(ctx, seedClient, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "calico-monitoring-config", Namespace: network.Namespace}}); err != nil {
			return fmt.Errorf("failed deleting calico-monitoring-config ConfigMap: %w", err)
		}
	}

	return calicoControlPlaneMonitoringChart.Apply(ctx, chartApplier, network.Namespace, nil, "", "", map[string]interface{}{"gep19Monitoring": gep19Monitoring})
}

// Reconcile implements Network.Actuator.
func (a *actuator) Reconcile(ctx context.Context, _ logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	var (
		networkConfig *calicov1alpha1.NetworkConfig
		err           error
	)
	ipFamilies := sets.New[extensionsv1alpha1.IPFamily](network.Spec.IPFamilies...)

	if network.Spec.ProviderConfig != nil {
		networkConfig, err = calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
		if err != nil {
			return err
		}
	}

	if networkConfig == nil {
		networkConfig = &calicov1alpha1.NetworkConfig{}
	}

	if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv4) {
		if networkConfig.IPv4 == nil {
			networkConfig.IPv4 = &calicov1alpha1.IPv4{}
		}
	}
	if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv6) {
		if networkConfig.IPv6 == nil {
			networkConfig.IPv6 = &calicov1alpha1.IPv6{}
		}
	}

	if cluster.Shoot.Spec.Networking != nil && cluster.Shoot.Spec.Networking.Nodes != nil && len(*cluster.Shoot.Spec.Networking.Nodes) > 0 {
		autodetectionMode := fmt.Sprintf("cidr=%s", *cluster.Shoot.Spec.Networking.Nodes)
		setAutoDetectionMethod(networkConfig, ipFamilies, autodetectionMode)

		if cluster.Shoot.Status.Networking != nil && cluster.Shoot.Status.Networking.Nodes != nil && len(cluster.Shoot.Status.Networking.Nodes) > 0 {
			ipv4Nodes, ipv6Nodes, err := segregateNodeCIDRs(cluster.Shoot.Status.Networking.Nodes)
			if err != nil {
				return err
			}

			autodetectionMode = updateAutoDetectionMode(ipv4Nodes)
			setAutoDetectionMethod(networkConfig, ipFamilies, autodetectionMode)

			autodetectionModeV6 := updateAutoDetectionMode(ipv6Nodes)
			setAutoDetectionMethodV6(networkConfig, ipFamilies, autodetectionModeV6)
		}
	}

	if networkConfig != nil && networkConfig.Overlay != nil {
		if networkConfig.Overlay.Enabled {
			if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv4) {
				networkConfig.IPv4.Mode = (*calicov1alpha1.PoolMode)(pointer.String(string(calicov1alpha1.Always)))
			}
			if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv6) {
				networkConfig.IPv6.Mode = (*calicov1alpha1.PoolMode)(pointer.String(string(calicov1alpha1.Always)))
			}
			networkConfig.Backend = (*calicov1alpha1.Backend)(pointer.String(string(calicov1alpha1.Bird)))
		} else {
			if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv4) {
				networkConfig.IPv4.Mode = (*calicov1alpha1.PoolMode)(pointer.String(string(calicov1alpha1.Never)))
			}
			if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv6) {
				networkConfig.IPv6.Mode = (*calicov1alpha1.PoolMode)(pointer.String(string(calicov1alpha1.Never)))
			}
			if networkConfig.Overlay.CreatePodRoutes != nil && *networkConfig.Overlay.CreatePodRoutes {
				networkConfig.Backend = (*calicov1alpha1.Backend)(pointer.String(string(calicov1alpha1.Bird)))
			} else {
				networkConfig.Backend = (*calicov1alpha1.Backend)(pointer.String(string(calicov1alpha1.None)))
			}
		}
	}

	if cluster.Shoot.Spec.Kubernetes.KubeProxy != nil && cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled != nil && !*cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled {
		if networkConfig == nil || networkConfig.EbpfDataplane == nil || (networkConfig.EbpfDataplane != nil && !networkConfig.EbpfDataplane.Enabled) {
			return field.Forbidden(field.NewPath("spec", "kubernetes", "kubeProxy", "enabled"), "Disabling kube-proxy is forbidden in conjunction with calico without running in ebpf dataplane")
		}
	}

	kubeProxyEnabled := true
	if cluster.Shoot.Spec.Kubernetes.KubeProxy != nil && cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled != nil {
		kubeProxyEnabled = *cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled
	}

	// Create shoot chart renderer
	chartRenderer, err := a.chartRendererFactory.NewChartRendererForShoot(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("could not create chart renderer for shoot '%s': %w", network.Namespace, err)
	}

	var podCIDRs []string
	if cluster.Shoot.Status.Networking != nil {
		podCIDRs = cluster.Shoot.Status.Networking.Pods
	}

	calicoChart, err := chartspkg.RenderCalicoChart(
		chartRenderer,
		network,
		networkConfig,
		cluster.Shoot.Spec.Kubernetes.Version,
		gardencorev1beta1helper.ShootWantsVerticalPodAutoscaler(cluster.Shoot),
		kubeProxyEnabled,
		features.FeatureGate.Enabled(features.NonPrivilegedCalicoNode),
		cluster.Shoot.Spec.Networking.Nodes,
		podCIDRs,
	)
	if err != nil {
		return err
	}

	data := map[string][]byte{chartspkg.CalicoConfigKey: calicoChart}
	if err := managedresources.CreateForShoot(ctx, a.client, network.Namespace, CalicoConfigManagedResourceName, "extension-networking-calico", false, data); err != nil {
		return err
	}

	if err := applyMonitoringConfig(ctx, a.client, a.chartApplier, network, false); err != nil {
		return err
	}

	return a.updateProviderStatus(ctx, network, networkConfig)
}

func setAutoDetectionMethod(networkConfig *calicov1alpha1.NetworkConfig, ipFamilies sets.Set[extensionsv1alpha1.IPFamily], autodetectionMode string) {
	if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv4) {
		networkConfig.IPv4.AutoDetectionMethod = &autodetectionMode
	}
}

func setAutoDetectionMethodV6(networkConfig *calicov1alpha1.NetworkConfig, ipFamilies sets.Set[extensionsv1alpha1.IPFamily], autodetectionModeV6 string) {
	if ipFamilies.Has(extensionsv1alpha1.IPFamilyIPv6) {
		networkConfig.IPv6.AutoDetectionMethod = &autodetectionModeV6
	}
}

func segregateNodeCIDRs(nodeCIDRs []string) ([]string, []string, error) {
	var ipv4Nodes, ipv6Nodes []string
	for _, nodeCidr := range nodeCIDRs {
		_, cidr, err := net.ParseCIDR(nodeCidr)
		if err != nil {
			return nil, nil, err
		}
		if cidr.IP.To4() != nil {
			ipv4Nodes = append(ipv4Nodes, nodeCidr)
		} else {
			ipv6Nodes = append(ipv6Nodes, nodeCidr)
		}
	}
	return ipv4Nodes, ipv6Nodes, nil
}

func updateAutoDetectionMode(nodes []string) string {
	if len(nodes) > 0 {
		return fmt.Sprintf("cidr=%s", strings.Join(nodes, ","))
	}
	return ""
}
