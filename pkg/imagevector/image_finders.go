// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package imagevector

import (
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"
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

// CalicoFlexVolumeDriverImage returns the Calico flexvol image.
func CalicoFlexVolumeDriverImage(kubernetesVersion string) string {
	return findImage(calico.PodToDaemonFlexVolumeDriverImageName, kubernetesVersion)
}

// ClusterProportionalAutoscalerImage returns the Calico cluster-proportional-autoscaler image.
func ClusterProportionalAutoscalerImage(kubernetesVersion string) string {
	return findImage(calico.CalicoClusterProportionalAutoscalerImageName, kubernetesVersion)
}

// ClusterProportionalVerticalAutoscalerImage returns the Calico cluster-proportional-vertical-autoscaler image.
func ClusterProportionalVerticalAutoscalerImage(kubernetesVersion string) string {
	return findImage(calico.ClusterProportionalVerticalAutoscalerImageName, kubernetesVersion)
}
