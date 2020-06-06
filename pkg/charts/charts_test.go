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

package charts_test

import (
	"fmt"

	mockchartrenderer "github.com/gardener/gardener/pkg/mock/gardener/chartrenderer"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	"github.com/gardener/gardener-extension-networking-calico/pkg/charts"
	"github.com/gardener/gardener-extension-networking-calico/pkg/imagevector"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/manifest"
)

var (
	trueVar    = true
	falseVar   = false
	mtuVar     = "1430"
	defaultMtu = "1440"
)

var _ = Describe("Chart package test", func() {
	var (
		podCIDR                                         = calicov1alpha1.CIDR("12.0.0.0/8")
		crossSubnet                                     = calicov1alpha1.CrossSubnet
		always                                          = calicov1alpha1.Always
		never                                           = calicov1alpha1.Never
		invalid             calicov1alpha1.IPv4PoolMode = "invalid"
		autodetectionMethod                             = "interface=eth1"
		backendNone                                     = calicov1alpha1.None
		backendVXLan                                    = calicov1alpha1.VXLan
		backendBird                                     = calicov1alpha1.Bird
		backendInvalid                                  = calicov1alpha1.Backend("invalid")
		poolIPIP                                        = calicov1alpha1.PoolIPIP
		poolVXlan                                       = calicov1alpha1.PoolVXLan

		network                  *extensionsv1alpha1.Network
		networkConfigNil         *calicov1alpha1.NetworkConfig
		networkConfigBackendNone *calicov1alpha1.NetworkConfig
		networkConfigAll         *calicov1alpha1.NetworkConfig
		networkConfigAllMTU      *calicov1alpha1.NetworkConfig
		networkConfigDeprecated  *calicov1alpha1.NetworkConfig
		networkConfigInvalid     *calicov1alpha1.NetworkConfig

		objectMeta = metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		}
	)

	BeforeEach(func() {
		network = &extensionsv1alpha1.Network{
			ObjectMeta: objectMeta,
			Spec: extensionsv1alpha1.NetworkSpec{
				ServiceCIDR: "10.0.0.0/8",
				PodCIDR:     string(podCIDR),
			},
		}
		networkConfigNil = nil
		networkConfigBackendNone = &calicov1alpha1.NetworkConfig{
			Backend: &backendNone,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
		}
		networkConfigAll = &calicov1alpha1.NetworkConfig{
			Backend: &backendVXLan,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPv4: &calicov1alpha1.IPv4{
				Pool:                &poolVXlan,
				Mode:                &crossSubnet,
				AutoDetectionMethod: &autodetectionMethod,
			},
		}
		networkConfigAllMTU = &calicov1alpha1.NetworkConfig{
			Backend: &backendVXLan,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPv4: &calicov1alpha1.IPv4{
				Pool:                &poolVXlan,
				Mode:                &crossSubnet,
				AutoDetectionMethod: &autodetectionMethod,
			},
			VethMTU: &mtuVar,
		}
		networkConfigDeprecated = &calicov1alpha1.NetworkConfig{
			Backend: &backendBird,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPIP:                  &crossSubnet,
			IPAutoDetectionMethod: &autodetectionMethod,
		}
		networkConfigInvalid = &calicov1alpha1.NetworkConfig{
			Backend: &backendInvalid,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPv4: &calicov1alpha1.IPv4{
				Mode:                &invalid,
				AutoDetectionMethod: &autodetectionMethod,
			},
		}
	})

	Describe("#ComputeCalicoChartValues", func() {
		It("empty network config should properly render calico chart values", func() {
			values, err := charts.ComputeCalicoChartValues(network, networkConfigNil)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(calicov1alpha1.Bird),
					"ipam": map[string]interface{}{
						"type":   "host-local",
						"subnet": "usePodCidr",
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": trueVar,
					},
					"veth_mtu": defaultMtu,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": true,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(poolIPIP),
						"mode":                string(always),
						"autoDetectionMethod": nil,
					},
				},
			}))
		})
	})

	Describe("#ComputeCalicoChartValues", func() {
		It("should disable felix ip in ip and set pool mode to never when setting backend to none", func() {
			values, err := charts.ComputeCalicoChartValues(network, networkConfigBackendNone)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(*networkConfigBackendNone.Backend),
					"ipam": map[string]interface{}{
						"type":   networkConfigBackendNone.IPAM.Type,
						"subnet": string(*networkConfigBackendNone.IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": falseVar,
					},
					"veth_mtu": defaultMtu,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": false,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(poolIPIP),
						"mode":                string(never),
						"autoDetectionMethod": nil,
					},
				},
			}))
		})
	})

	Describe("#ComputeAllCalicoChartValues", func() {
		It("should correctly compute all of the calico chart values", func() {
			values, err := charts.ComputeCalicoChartValues(network, networkConfigAll)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(*networkConfigAll.Backend),
					"ipam": map[string]interface{}{
						"type":   networkConfigAll.IPAM.Type,
						"subnet": string(*networkConfigAll.IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": trueVar,
					},
					"veth_mtu": defaultMtu,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": true,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(poolVXlan),
						"mode":                string(*networkConfigAll.IPv4.Mode),
						"autoDetectionMethod": *networkConfigAll.IPv4.AutoDetectionMethod,
					},
				},
			}))
		})
		It("should correctly compute all of the calico chart values with mtu", func() {
			values, err := charts.ComputeCalicoChartValues(network, networkConfigAllMTU)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(*networkConfigAll.Backend),
					"ipam": map[string]interface{}{
						"type":   networkConfigAll.IPAM.Type,
						"subnet": string(*networkConfigAll.IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": trueVar,
					},
					"veth_mtu": mtuVar,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": true,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(poolVXlan),
						"mode":                string(*networkConfigAll.IPv4.Mode),
						"autoDetectionMethod": *networkConfigAll.IPv4.AutoDetectionMethod,
					},
				},
			}))
		})
	})

	Describe("#ComputeAllCalicoChartValues", func() {
		It("should respect deprecated fields in order to keep backwards compatibility", func() {
			values, err := charts.ComputeCalicoChartValues(network, networkConfigDeprecated)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(*networkConfigDeprecated.Backend),
					"ipam": map[string]interface{}{
						"type":   networkConfigDeprecated.IPAM.Type,
						"subnet": string(*networkConfigDeprecated.IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": trueVar,
					},
					"veth_mtu": defaultMtu,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": true,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(calicov1alpha1.PoolIPIP),
						"mode":                string(*networkConfigDeprecated.IPIP),
						"autoDetectionMethod": *networkConfigDeprecated.IPAutoDetectionMethod,
					},
				},
			}))
		})
	})

	Describe("#ComputeInvalidCalicoChartValues", func() {
		It("should error on invalid config value", func() {
			_, err := charts.ComputeCalicoChartValues(network, networkConfigInvalid)
			Expect(err).To(Equal(fmt.Errorf("error when generating calico config: unsupported value for backend: invalid")))
		})
	})

	Describe("#RenderCalicoChart", func() {
		var (
			ctrl                = gomock.NewController(GinkgoT())
			mockChartRenderer   = mockchartrenderer.NewMockInterface(ctrl)
			testManifestContent = "test-content"
			mkManifest          = func(name string) manifest.Manifest {
				return manifest.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		)
		It("Render Calico charts correctly", func() {
			mockChartRenderer.EXPECT().Render(calico.ChartPath, calico.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []manifest.Manifest{
					mkManifest(charts.CalicoConfigKey),
				},
			}, nil)

			_, err := charts.RenderCalicoChart(mockChartRenderer, network, networkConfigNil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#OverrideImages", func() {
		override := "override"
		defaults := func(key, value string) map[string]interface{} {
			def := map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
					"calico-podtodaemon-flex": imagevector.CalicoFlexVolumeDriverImage(),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"config": map[string]interface{}{
					"backend": string(*networkConfigBackendNone.Backend),
					"ipam": map[string]interface{}{
						"type":   networkConfigBackendNone.IPAM.Type,
						"subnet": string(*networkConfigBackendNone.IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": falseVar,
					},
					"veth_mtu": defaultMtu,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": false,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                string(poolIPIP),
						"mode":                string(never),
						"autoDetectionMethod": nil,
					},
				},
			}
			def["images"].(map[string]interface{})[key] = value
			return def
		}

		It("should respect override selected images (cni)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				CNIImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-cni", "override")))
		})
		It("should respect override selected images (node)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				NodeImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-node", "override")))
		})
		It("should respect override selected images (kube controllers)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				KubeControllersImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-kube-controllers", "override")))
		})
		It("should respect override selected images (podtodaemon-flex)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				PodToDaemonFlexVolumeDriverImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-podtodaemon-flex", "override")))
		})
		It("should respect override selected images (typha)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				TyphaImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-typha", "override")))
		})
		It("should respect override selected images (cpa)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				CalicoClusterProportionalAutoscalerImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-cpa", "override")))
		})
		It("should respect override selected images (cpva)", func() {
			config := *networkConfigBackendNone
			config.Images = &calicov1alpha1.Images{
				ClusterProportionalVerticalAutoscalerImageName: &override,
			}
			values, err := charts.ComputeCalicoChartValues(network, &config)
			Expect(err).To(BeNil())
			Expect(values).To(Equal(defaults("calico-cpva", "override")))
		})
	})
})
