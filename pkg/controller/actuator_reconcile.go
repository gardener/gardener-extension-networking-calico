// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"

	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/apis/extensions/validation"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	"github.com/labstack/gommon/log"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

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

	if network.Spec.ProviderConfig != nil && network.Spec.ProviderConfig.Raw != nil {
		networkConfig, err := calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
		if err != nil {
			return err
		}
		if networkConfig.BirdExporter != nil && networkConfig.BirdExporter.Enabled {
			calicoControlPlaneMonitoringChart.Objects = append(calicoControlPlaneMonitoringChart.Objects, &chart.Object{
				Type: &monitoringv1alpha1.ScrapeConfig{},
				Name: "shoot-calico-bird",
			})
		}
	}

	if deleteChart {
		return client.IgnoreNotFound(calicoControlPlaneMonitoringChart.Delete(ctx, seedClient, network.Namespace))
	}

	return calicoControlPlaneMonitoringChart.Apply(ctx, chartApplier, network.Namespace, nil, "", "", nil)
}

// Reconcile implements Network.Actuator.
func (a *actuator) Reconcile(ctx context.Context, _ logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	if errList := validation.ValidateNetwork(network); len(errList) != 0 {
		return fmt.Errorf("invalid network resource: %w", errList.ToAggregate())
	}
	var (
		networkConfig *calicov1alpha1.NetworkConfig
		err           error
	)

	ipFamilies := slices.Clone(network.Spec.IPFamilies)

	if network.Spec.ProviderConfig != nil {
		networkConfig, err = calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
		if err != nil {
			return err
		}

		if err := ValidateNetworkConfig(networkConfig); err != nil {
			return err
		}

	}

	if condition := gardencorev1beta1helper.GetCondition(cluster.Shoot.Status.Constraints, v1beta1.ShootDualStackNodesMigrationReady); condition != nil && condition.Status != v1beta1.ConditionTrue {
		if len(ipFamilies) > 1 {
			ipFamilies = ipFamilies[:1]
		}
	}

	if networkConfig == nil {
		networkConfig = &calicov1alpha1.NetworkConfig{}
	}

	if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv4) {
		if networkConfig.IPv4 == nil {
			networkConfig.IPv4 = &calicov1alpha1.IPv4{}
		}
	}
	if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv6) {
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

	overlaySwitch, err := isOverlaySwitch(ctx, a.client, network)
	if err != nil {
		return err
	}

	// Only check node routes if overlay switch is happening and feature flag is enabled
	if overlaySwitch && features.FeatureGate.Enabled(features.SeamlessOverlaySwitch) {
		// Try to get shoot client to check node conditions
		shootClient, err := a.getShootClient(ctx, cluster)
		if err != nil {
			// Cannot access shoot cluster - we must wait before switching overlay off
			log.Info("Cannot access shoot cluster to verify node routes - waiting", "error", err)
			return fmt.Errorf("cannot verify node routes before overlay switch: %w", err)
		}

		allNodesRoutesCreated, err := areAllNodesRoutesCreated(ctx, shootClient)
		if err != nil {
			log.Info("Failed to check node route status - waiting", "error", err)
			return fmt.Errorf("failed to check node route status: %w", err)
		}

		log.Info("Overlay switch check",
			"overlaySwitch: ", overlaySwitch,
			"allNodesRoutesCreated: ", allNodesRoutesCreated)

		// If routes are not ready, keep overlay enabled
		if !allNodesRoutesCreated {
			if networkConfig.Overlay == nil {
				networkConfig.Overlay = &calicov1alpha1.Overlay{}
			}
			networkConfig.Overlay.Enabled = true
			log.Info("Forcing overlay to remain enabled - waiting for routes to be created on all nodes")
			return fmt.Errorf("waiting for routes to be created on all nodes before disabling overlay")
		}
	}

	if networkConfig != nil {
		if networkConfig.Overlay != nil {
			if networkConfig.Overlay.Enabled {
				setPoolMode(networkConfig, ipFamilies, calicov1alpha1.Always)
				if networkConfig.VXLAN != nil && networkConfig.VXLAN.Enabled {
					networkConfig.Backend = (*calicov1alpha1.Backend)(ptr.To(string(calicov1alpha1.VXLan)))
				}
			} else {
				setPoolMode(networkConfig, ipFamilies, calicov1alpha1.Never)
				if networkConfig.Overlay.CreatePodRoutes != nil && *networkConfig.Overlay.CreatePodRoutes {
					networkConfig.Backend = (*calicov1alpha1.Backend)(ptr.To(string(calicov1alpha1.Bird)))
				}
			}
		} else {
			if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv6) {
				setPoolMode(networkConfig, ipFamilies, calicov1alpha1.Never)
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
		ipFamilies,
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

func setPoolMode(networkConfig *calicov1alpha1.NetworkConfig, ipFamilies []extensionsv1alpha1.IPFamily, mode calicov1alpha1.PoolMode) {
	if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv6) {
		networkConfig.IPv6.Mode = (*calicov1alpha1.PoolMode)(ptr.To(string(mode)))
	} else {
		networkConfig.IPv4.Mode = (*calicov1alpha1.PoolMode)(ptr.To(string(mode)))
	}

	if mode == calicov1alpha1.Never {
		networkConfig.Backend = (*calicov1alpha1.Backend)(ptr.To(string(calicov1alpha1.None)))
	} else {
		networkConfig.Backend = (*calicov1alpha1.Backend)(ptr.To(string(calicov1alpha1.Bird)))
	}
}

func setAutoDetectionMethod(networkConfig *calicov1alpha1.NetworkConfig, ipFamilies []extensionsv1alpha1.IPFamily, autodetectionMode string) {
	if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv4) {
		networkConfig.IPv4.AutoDetectionMethod = &autodetectionMode
	}
}

func setAutoDetectionMethodV6(networkConfig *calicov1alpha1.NetworkConfig, ipFamilies []extensionsv1alpha1.IPFamily, autodetectionModeV6 string) {
	if slices.Contains(ipFamilies, extensionsv1alpha1.IPFamilyIPv6) {
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

// getShootClient creates a client for the shoot cluster
func (a *actuator) getShootClient(ctx context.Context, cluster *extensionscontroller.Cluster) (client.Client, error) {
	_, shootClient, err := util.NewClientForShoot(ctx, a.client, cluster.ObjectMeta.Name, client.Options{}, extensionsconfig.RESTOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create shoot client: %w", err)
	}
	return shootClient, nil
}

// areAllNodesRoutesCreated checks if all nodes in the shoot cluster have the NetworkUnavailable
// condition set to False with reason RouteCreated
func areAllNodesRoutesCreated(ctx context.Context, shootClient client.Client) (bool, error) {
	nodeList := &corev1.NodeList{}
	if err := shootClient.List(ctx, nodeList); err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
	}

	// If there are no nodes yet, we cannot proceed
	if len(nodeList.Items) == 0 {
		return false, nil
	}

	// Check each node for the NetworkUnavailable condition
	for _, node := range nodeList.Items {
		hasRouteCreated := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeNetworkUnavailable {
				// NetworkUnavailable should be False and reason should be RouteCreated
				if condition.Status == corev1.ConditionFalse && condition.Reason == "RouteCreated" {
					hasRouteCreated = true
					break
				}
			}
		}
		if !hasRouteCreated {
			log.Info("Node does not have route created yet", "node", node.Name)
			return false, nil
		}
	}

	return true, nil
}

func isOverlaySwitch(ctx context.Context, seedClient client.Client, network *extensionsv1alpha1.Network) (bool, error) {
	shootOverlayEnabled := true
	networkConfig, err := calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
	if err != nil {
		return false, err
	}

	if networkConfig.Overlay != nil {
		shootOverlayEnabled = networkConfig.Overlay.Enabled
	}

	// Get the calico-node DaemonSet from the ManagedResource
	calicoDaemonSet, err := getDaemonSetFromManagedResource(ctx, seedClient, network.Namespace, CalicoConfigManagedResourceName, "calico-node")
	if err != nil {
		// ManagedResource doesn't exist or DaemonSet not found
		// If the shoot wants overlay disabled, treat this as a potential switch
		// to be safe and wait for RouteController
		if !shootOverlayEnabled {
			log.Info("Cannot read current overlay state, but shoot wants overlay disabled - treating as overlay switch", "error", err)
			return true, nil
		}
		// If shoot wants overlay enabled or first reconciliation of a new cluster happens, no switch
		log.Info("Cannot read current overlay state during first reconciliation or with overlay enabled", "error", err)
		return false, nil
	}

	overlayEnabled := false
	for _, container := range calicoDaemonSet.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == "FELIX_IPINIPENABLED" && env.Value == "true" || env.Name == "FELIX_VXLANENABLED" && env.Value == "true" {
				overlayEnabled = true
			}
		}
	}

	if shootOverlayEnabled != overlayEnabled {
		return true, nil
	}

	return false, nil
}

// getDaemonSetFromManagedResource extracts a specific DaemonSet from a ManagedResource's secret
// It only decodes DaemonSet objects to avoid scheme registration issues with other resource types
func getDaemonSetFromManagedResource(ctx context.Context, c client.Client, namespace, mrName, daemonSetName string) (*appsv1.DaemonSet, error) {
	// Get the ManagedResource
	managedResource := &resourcesv1alpha1.ManagedResource{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: mrName}, managedResource); err != nil {
		return nil, fmt.Errorf("could not get ManagedResource %q: %w", mrName, err)
	}

	// Create a scheme with only the types we need to decode
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("could not add apps/v1 to scheme: %w", err)
	}
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	// Iterate through all secrets referenced by the ManagedResource
	for _, secretRef := range managedResource.Spec.SecretRefs {
		secret := &corev1.Secret{}
		if err := c.Get(ctx, client.ObjectKey{Name: secretRef.Name, Namespace: namespace}, secret); err != nil {
			return nil, fmt.Errorf("could not get secret %q: %w", secretRef.Name, err)
		}

		// Check all keys in the secret data
		for _, value := range secret.Data {
			// Split the YAML content into individual documents
			docs := strings.Split(string(value), "---\n")

			for _, doc := range docs {
				if strings.TrimSpace(doc) == "" {
					continue
				}

				// Try to parse as unstructured to check if it's a DaemonSet
				var meta struct {
					Kind     string `yaml:"kind"`
					Metadata struct {
						Name string `yaml:"name"`
					} `yaml:"metadata"`
				}

				if err := yaml.Unmarshal([]byte(doc), &meta); err != nil {
					// Skip documents that can't be parsed as YAML
					continue
				}

				// Only decode if it's a DaemonSet with the name we're looking for
				if meta.Kind == "DaemonSet" && meta.Metadata.Name == daemonSetName {
					obj, _, err := decoder.Decode([]byte(doc), nil, nil)
					if err != nil {
						return nil, fmt.Errorf("could not decode DaemonSet %q: %w", daemonSetName, err)
					}

					if ds, ok := obj.(*appsv1.DaemonSet); ok {
						return ds, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("DaemonSet %q not found in ManagedResource %q", daemonSetName, mrName)
}
