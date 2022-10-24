// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	calicov1alpha1helper "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1/helper"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	"github.com/gardener/gardener-extension-networking-calico/pkg/charts"
	"github.com/gardener/gardener-extension-networking-calico/pkg/features"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/chart"
	"github.com/gardener/gardener/pkg/utils/managedresources/builder"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// CalicoConfigSecretName is the name of the secret used for the managed resource of networking calico
	CalicoConfigSecretName = "extension-networking-calico-config"
)

func withLocalObjectRefs(refs ...string) []corev1.LocalObjectReference {
	var localObjectRefs []corev1.LocalObjectReference
	for _, ref := range refs {
		localObjectRefs = append(localObjectRefs, corev1.LocalObjectReference{Name: ref})
	}
	return localObjectRefs
}

func calicoSecret(cl client.Client, calicoConfig []byte, namespace string) (*builder.Secret, []corev1.LocalObjectReference) {
	return builder.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.CalicoConfigKey: calicoConfig}).
		WithNamespacedName(namespace, CalicoConfigSecretName), withLocalObjectRefs(CalicoConfigSecretName)
}

func activateSystemComponentsNodeSelector(shoot *gardencorev1beta1.Shoot) bool {
	var atLeastOneWorkerPoolHasSystemComponents bool

	for _, worker := range shoot.Spec.Provider.Workers {
		if gardencorev1beta1helper.SystemComponentsAllowed(&worker) {
			atLeastOneWorkerPoolHasSystemComponents = true
			break
		}
	}

	return atLeastOneWorkerPoolHasSystemComponents
}

func applyMonitoringConfig(ctx context.Context, seedClient client.Client, chartApplier gardenerkubernetes.ChartApplier, network *extensionsv1alpha1.Network, deleteChart bool) error {
	calicoControlPlaneMonitoringChart := &chart.Chart{
		Name: calico.MonitoringName,
		Path: calico.CalicoMonitoringChartPath,
		Objects: []*chart.Object{
			{
				Type: &corev1.ConfigMap{},
				Name: calico.MonitoringName,
			},
		},
	}

	if deleteChart {
		return client.IgnoreNotFound(calicoControlPlaneMonitoringChart.Delete(ctx, seedClient, network.Namespace))
	}

	return calicoControlPlaneMonitoringChart.Apply(ctx, chartApplier, network.Namespace, nil, "", "", nil)
}

// Reconcile implements Network.Actuator.
func (a *actuator) Reconcile(ctx context.Context, _ logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	var (
		networkConfig *calicov1alpha1.NetworkConfig
		err           error
	)

	if network.Spec.ProviderConfig != nil {
		networkConfig, err = calicov1alpha1helper.CalicoNetworkConfigFromNetworkResource(network)
		if err != nil {
			return err
		}
	}

	if cluster.Shoot.Spec.Networking.Nodes != nil && len(*cluster.Shoot.Spec.Networking.Nodes) > 0 {
		autodetectionMode := fmt.Sprintf("cidr=%s", *cluster.Shoot.Spec.Networking.Nodes)
		if networkConfig == nil {
			networkConfig = &calicov1alpha1.NetworkConfig{}
		}
		if networkConfig.IPv4 == nil {
			networkConfig.IPv4 = &calicov1alpha1.IPv4{}
		}
		networkConfig.IPv4.AutoDetectionMethod = &autodetectionMode

		if networkConfig.IPv6 == nil {
			autoNone := "none"
			autoKubeInt := "kubernetes-internal-ip"
			networkConfig.IPv6 = &calicov1alpha1.IPv6{}
			networkConfig.IPv6.AutoDetectionMethod = &autoKubeInt

			// Set IPv4 autodetect to none when using IPv6
			networkConfig.IPv4.AutoDetectionMethod = &autoNone
		}
	}

	if cluster.Shoot.Spec.Kubernetes.KubeProxy != nil && cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled != nil && !*cluster.Shoot.Spec.Kubernetes.KubeProxy.Enabled {
		if networkConfig.EbpfDataplane == nil || (networkConfig.EbpfDataplane != nil && !networkConfig.EbpfDataplane.Enabled) {
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

	calicoChart, err := charts.RenderCalicoChart(
		chartRenderer,
		network,
		networkConfig,
		activateSystemComponentsNodeSelector(cluster.Shoot),
		cluster.Shoot.Spec.Kubernetes.Version,
		gardencorev1beta1helper.ShootWantsVerticalPodAutoscaler(cluster.Shoot),
		kubeProxyEnabled,
		gardencorev1beta1helper.IsPSPDisabled(cluster.Shoot),
		features.FeatureGate.Enabled(features.NonPrivilegedCalicoNode),
	)
	if err != nil {
		return err
	}

	secret, secretRefs := calicoSecret(a.client, calicoChart, network.Namespace)
	if err := secret.Reconcile(ctx); err != nil {
		return err
	}

	if err := builder.
		NewManagedResource(a.client).
		WithNamespacedName(network.Namespace, CalicoConfigSecretName).
		WithSecretRefs(secretRefs).
		WithInjectedLabels(map[string]string{constants.ShootNoCleanup: "true"}).
		Reconcile(ctx); err != nil {
		return err
	}

	if err := applyMonitoringConfig(ctx, a.client, a.chartApplier, network, false); err != nil {
		return err
	}

	return a.updateProviderStatus(ctx, network, networkConfig)
}
