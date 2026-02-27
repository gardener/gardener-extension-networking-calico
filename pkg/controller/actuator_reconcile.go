// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	extensionsconfig "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	"github.com/gardener/gardener/pkg/api/extensions/validation"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	versionutils "github.com/gardener/gardener/pkg/utils/version"
	"github.com/go-logr/logr"
	monitoringv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/utils/ptr"
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

var (
	daemonSetScheme  = runtime.NewScheme()
	daemonSetDecoder runtime.Decoder
)

func init() {
	utilruntime.Must(appsv1.AddToScheme(daemonSetScheme))
	daemonSetDecoder = serializer.NewCodecFactory(daemonSetScheme).UniversalDeserializer()
}

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
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
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

	shootKubernetesVersion, err := semver.NewVersion(cluster.Shoot.Spec.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("failed to parse shoot Kubernetes version %s: %w", cluster.Shoot.Spec.Kubernetes.Version, err)
	}

	newK8sGreaterEqual136, err := versionutils.CheckVersionMeetsConstraint(shootKubernetesVersion.String(), ">= 1.36")
	if err != nil {
		return fmt.Errorf("failed to check version constraint %w", err)
	}

	if features.FeatureGate.Enabled(features.SeamlessOverlaySwitch) && (newK8sGreaterEqual136 || isMutatingAdmissionPolicyEnabled(cluster)) {
		overlaySwitch, err := isOverlaySwitch(ctx, log, a.client, network)
		if err != nil {
			return fmt.Errorf("failed to detect pod overlay switch: %w", err)
		}

		if err := a.ensureNodesRoutesBeforeOverlaySwitch(ctx, log, cluster, networkConfig, overlaySwitch); err != nil {
			return err
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

// ensureNodesRoutesBeforeOverlaySwitch checks if all nodes have routes created before allowing overlay to be disabled
func (a *actuator) ensureNodesRoutesBeforeOverlaySwitch(ctx context.Context, log logr.Logger, cluster *extensionscontroller.Cluster, networkConfig *calicov1alpha1.NetworkConfig, overlaySwitch bool) error {
	if !overlaySwitch {
		return nil
	}

	shootClient, err := a.getShootClient(ctx, cluster)
	if err != nil {
		// Cannot access shoot cluster - we must wait before switching overlay off
		return fmt.Errorf("cannot verify node routes before overlay switch: %w", err)
	}

	allNodeRoutesCreated, err := areAllNodeRoutesCreated(ctx, log, shootClient)
	if err != nil {
		return fmt.Errorf("failed to check node route status: %w", err)
	}

	// If routes are not ready, return error to retry later
	if !allNodeRoutesCreated {
		return fmt.Errorf("waiting for routes to be created on all nodes before disabling overlay")
	}

	return nil
}

// areAllNodeRoutesCreated checks if all nodes in the shoot cluster have the NetworkUnavailable
// condition set to False with reason RouteCreated
func areAllNodeRoutesCreated(ctx context.Context, log logr.Logger, shootClient client.Client) (bool, error) {
	nodeList := &corev1.NodeList{}
	if err := shootClient.List(ctx, nodeList); err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Check each node for the NetworkUnavailable condition
	for _, node := range nodeList.Items {
		if !hasRouteCreated(node) {
			log.Info("Node does not have route created yet", "nodeName", node.Name)
			return false, nil
		}
	}

	return true, nil
}

// hasRouteCreated checks if a node has the NetworkUnavailable condition set to False with reason RouteCreated
func hasRouteCreated(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeNetworkUnavailable {
			if condition.Status == corev1.ConditionFalse && condition.Reason == "RouteCreated" {
				return true
			}
		}
	}
	return false
}

// isOverlaySwitch determines if there is a switch from overlay to non-overlay networking based on the desired state in the Network resource and the actual state in the calico-node DaemonSet
func isOverlaySwitch(ctx context.Context, log logr.Logger, seedClient client.Client, network *extensionsv1alpha1.Network) (bool, error) {
	desiredOverlayEnabled := true

	if network.Spec.ProviderConfig != nil && network.Spec.ProviderConfig.Raw != nil {
		networkConfig, err := calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
		if err != nil {
			return false, fmt.Errorf("failed to decode network provider config: %w", err)
		}

		if networkConfig.Overlay != nil {
			desiredOverlayEnabled = networkConfig.Overlay.Enabled
		}
	}

	// Get the calico-node DaemonSet from the ManagedResource
	calicoDaemonSet, err := getDaemonSetFromManagedResource(ctx, seedClient, network.Namespace, CalicoConfigManagedResourceName, "calico-node")
	if err != nil {
		// If shoot wants overlay enabled or first reconciliation of a new cluster happens, no switch
		log.Info("Cannot read current overlay state during first reconciliation or with overlay enabled", "error", err)
		return false, nil
	}

	actualOverlayEnabled := false
	for _, container := range calicoDaemonSet.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == "FELIX_IPINIPENABLED" && env.Value == "true" || env.Name == "FELIX_VXLANENABLED" && env.Value == "true" {
				actualOverlayEnabled = true
			}
		}
	}

	return (desiredOverlayEnabled != actualOverlayEnabled) && !desiredOverlayEnabled, nil
}

// getDaemonSetFromManagedResource extracts a specific DaemonSet from a ManagedResource's secret
// It only decodes DaemonSet objects to avoid scheme registration issues with other resource types
func getDaemonSetFromManagedResource(ctx context.Context, c client.Client, namespace, mrName, daemonSetName string) (*appsv1.DaemonSet, error) {
	// Get the ManagedResource
	managedResource := &resourcesv1alpha1.ManagedResource{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: mrName}, managedResource); err != nil {
		return nil, fmt.Errorf("could not get ManagedResource %q: %w", mrName, err)
	}

	// Iterate through all secrets referenced by the ManagedResource
	for _, secretRef := range managedResource.Spec.SecretRefs {
		secret := &corev1.Secret{}
		if err := c.Get(ctx, client.ObjectKey{Name: secretRef.Name, Namespace: namespace}, secret); err != nil {
			return nil, fmt.Errorf("could not get secret %q: %w", secretRef.Name, err)
		}

		for _, value := range secret.Data {
			reader := yaml.NewYAMLReader(bufio.NewReader(strings.NewReader(string(value))))

			for {
				doc, err := reader.Read()
				if err != nil {
					break // End of document stream
				}

				// Decode the document to check its type
				obj, gvk, err := daemonSetDecoder.Decode(doc, nil, nil)
				if err != nil {
					continue // Skip documents that can't be decoded as DaemonSet
				}

				// Check if it's a DaemonSet with the name we're looking for
				if gvk.Kind == "DaemonSet" {
					if ds, ok := obj.(*appsv1.DaemonSet); ok && ds.Name == daemonSetName {
						return ds, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("DaemonSet %q not found in ManagedResource %q", daemonSetName, mrName)
}

func isMutatingAdmissionPolicyEnabled(cluster *extensionscontroller.Cluster) bool {
	if cluster.Shoot.Spec.Kubernetes.KubeAPIServer == nil {
		return false
	}

	if cluster.Shoot.Spec.Kubernetes.KubeAPIServer.FeatureGates == nil {
		return false
	}

	if enabled, ok := cluster.Shoot.Spec.Kubernetes.KubeAPIServer.FeatureGates["MutatingAdmissionPolicy"]; !ok || !enabled {
		return false
	}

	if cluster.Shoot.Spec.Kubernetes.KubeAPIServer.RuntimeConfig == nil {
		return false
	}

	if enabled, ok := cluster.Shoot.Spec.Kubernetes.KubeAPIServer.RuntimeConfig["admissionregistration.k8s.io/v1alpha1"]; !ok || !enabled {
		return false
	}

	return true
}
