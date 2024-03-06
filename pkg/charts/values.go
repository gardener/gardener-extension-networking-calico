// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package charts

import (
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/gardener-extension-networking-calico/charts"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

const CalicoConfigKey = "config.yaml"

// RenderCalicoChart renders the calico chart with the given values.
func RenderCalicoChart(
	renderer chartrenderer.Interface,
	network *extensionsv1alpha1.Network,
	config *calicov1alpha1.NetworkConfig,
	kubernetesVersion string,
	wantsVPA bool,
	kubeProxyEnabled bool,
	nonPrivileged bool,
	nodeCIDR *string,
) ([]byte, error) {
	values, err := ComputeCalicoChartValues(network, config, kubernetesVersion, wantsVPA, kubeProxyEnabled, nonPrivileged, nodeCIDR)
	if err != nil {
		return nil, err
	}
	release, err := renderer.RenderEmbeddedFS(charts.InternalChart, calico.CalicoChartPath, calico.ReleaseName, metav1.NamespaceSystem, values)
	if err != nil {
		return nil, err
	}

	return release.Manifest(), nil
}
