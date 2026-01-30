// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package imagevector

import (
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/version"

	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

func findImage(name, kubernetesVersion string) string {
	image, err := imageVector.FindImage(name, imagevector.RuntimeVersion(kubernetesVersion), imagevector.TargetVersion(kubernetesVersion))
	runtime.Must(err)
	return image.String()
}

// CalicoCNIImage returns the Calico CNI Image.
func CalicoCNIImage(kubernetesVersion string) string {
	return findImage(calico.CNIImageName, kubernetesVersion)
}

// CalicoNodeImage returns the Calico Node image.
func CalicoNodeImage(kubernetesVersion string) string {
	return findImage(calico.NodeImageName, kubernetesVersion)
}

// CalicoTyphaImage returns the Calico Typha image.
func CalicoTyphaImage(kubernetesVersion string) string {
	return findImage(calico.TyphaImageName, kubernetesVersion)
}

// CalicoKubeControllersImage returns the Calico Kube-controllers image.
func CalicoKubeControllersImage(kubernetesVersion string) string {
	return findImage(calico.KubeControllersImageName, kubernetesVersion)
}

// ClusterProportionalAutoscalerImage returns the Calico cluster-proportional-autoscaler image.
func ClusterProportionalAutoscalerImage(kubernetesVersion string) string {
	return findImage(calico.CalicoClusterProportionalAutoscalerImageName, kubernetesVersion)
}

// ClusterProportionalVerticalAutoscalerImage returns the Calico cluster-proportional-vertical-autoscaler image.
func ClusterProportionalVerticalAutoscalerImage(kubernetesVersion string) string {
	return findImage(calico.ClusterProportionalVerticalAutoscalerImageName, kubernetesVersion)
}

// BirdExporterImage returns the bird-exporter image.
func BirdExporterImage(kubernetesVersion string) string {
	return findImage(calico.BirdExporterImageName, kubernetesVersion)
}

// MultusImage returns the Multus CNI image.
func MultusImage(kubernetesVersion string) string {
	return findImage(calico.MultusImageName, kubernetesVersion)
}

// CNIPluginsImage returns the CNI plugins image.
func CNIPluginsImage(kubernetesVersion string) string {
	image, err := imageVector.FindImage(calico.CNIPluginsImageName)
	runtime.Must(err)
	image.WithOptionalTag(version.Get().GitVersion)
	return image.String()
}
