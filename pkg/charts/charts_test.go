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

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/chartrenderer"
	mockchartrenderer "github.com/gardener/gardener/pkg/chartrenderer/mock"
	"github.com/gardener/gardener/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/manifest"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	"github.com/gardener/gardener-extension-networking-calico/pkg/charts"
	"github.com/gardener/gardener-extension-networking-calico/pkg/imagevector"
)

var (
	trueVar    = true
	mtuVar     = "1430"
	defaultMtu = "1440"
)

var _ = Describe("Chart package test", func() {
	var (
		kubernetesVersion                               = "1.23.0"
		podCIDR                                         = calicov1alpha1.CIDR("12.0.0.0/8")
		nodeCIDR                                        = "10.250.0.0/8"
		usePodCidr                                      = calicov1alpha1.CIDR("usePodCidr")
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

		network                       *extensionsv1alpha1.Network
		networkConfigNil              *calicov1alpha1.NetworkConfig
		networkConfigNilValues        *calicov1alpha1.NetworkConfig
		networkConfigBackendNone      *calicov1alpha1.NetworkConfig
		networkConfigAll              *calicov1alpha1.NetworkConfig
		networkConfigAllMTU           *calicov1alpha1.NetworkConfig
		networkConfigAllEBPFDataplane *calicov1alpha1.NetworkConfig
		networkConfigDeprecated       *calicov1alpha1.NetworkConfig
		networkConfigInvalid          *calicov1alpha1.NetworkConfig
		networkConfigOverlayDisabled  *calicov1alpha1.NetworkConfig

		networkConfigNilFunc              = func() *calicov1alpha1.NetworkConfig { return networkConfigNil }
		networkConfigNilValuesFunc        = func() *calicov1alpha1.NetworkConfig { return networkConfigNilValues }
		networkConfigBackendNoneFunc      = func() *calicov1alpha1.NetworkConfig { return networkConfigBackendNone }
		networkConfigAllFunc              = func() *calicov1alpha1.NetworkConfig { return networkConfigAll }
		networkConfigAllMTUFunc           = func() *calicov1alpha1.NetworkConfig { return networkConfigAllMTU }
		networkConfigAllEBPFDataplaneFunc = func() *calicov1alpha1.NetworkConfig { return networkConfigAllEBPFDataplane }
		networkConfigDeprecatedFunc       = func() *calicov1alpha1.NetworkConfig { return networkConfigDeprecated }
		networkConfigOverlayDisabledFunc  = func() *calicov1alpha1.NetworkConfig { return networkConfigOverlayDisabled }

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
		networkConfigNilValues = &calicov1alpha1.NetworkConfig{
			Backend: &backendBird,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &usePodCidr,
				Type: "host-local",
			},
		}
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
		networkConfigAllEBPFDataplane = &calicov1alpha1.NetworkConfig{
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
			EbpfDataplane: &calicov1alpha1.EbpfDataplane{
				Enabled: true,
			},
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
		networkConfigOverlayDisabled = &calicov1alpha1.NetworkConfig{
			Overlay: &calicov1alpha1.Overlay{Enabled: false},
			Backend: &backendNone,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			IPv4: &calicov1alpha1.IPv4{
				Mode:                &never,
				AutoDetectionMethod: &autodetectionMethod,
			},
			IPAutoDetectionMethod: &autodetectionMethod,
		}
	})

	DescribeTable("#ComputeCalicoChartValues",
		func(config func() *calicov1alpha1.NetworkConfig, configResult func() *calicov1alpha1.NetworkConfig, wantsVPA bool,
			kubeProxyEnabled bool, isPSPDisabled bool, mtu string, ipinip bool, bpf bool, pool string,
			modeFunc func() string, detectionMethodFunc func() *string, nodesFunc func() *string, additionalGlobalOptions map[string]string) {
			values, err := charts.ComputeCalicoChartValues(network, config(), kubernetesVersion, wantsVPA, kubeProxyEnabled, isPSPDisabled, false, nodesFunc())
			Expect(err).To(BeNil())
			expected := map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(kubernetesVersion),
					"calico-typha":            imagevector.CalicoTyphaImage(kubernetesVersion),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(kubernetesVersion),
					"calico-node":             imagevector.CalicoNodeImage(kubernetesVersion),
					"calico-cpa":              imagevector.ClusterProportionalAutoscalerImage(kubernetesVersion),
					"calico-cpva":             imagevector.ClusterProportionalVerticalAutoscalerImage(kubernetesVersion),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
				"vpa": map[string]interface{}{
					"enabled": wantsVPA,
				},
				"config": map[string]interface{}{
					"backend": string(*configResult().Backend),
					"ipam": map[string]interface{}{
						"type":   configResult().IPAM.Type,
						"subnet": string(*configResult().IPAM.CIDR),
					},
					"typha": map[string]interface{}{
						"enabled": trueVar,
					},
					"kubeControllers": map[string]interface{}{
						"enabled": configResult().Backend != &backendNone,
					},
					"veth_mtu": mtu,
					"monitoring": map[string]interface{}{
						"enabled":          true,
						"typhaMetricsPort": "9093",
						"felixMetricsPort": "9091",
					},
					"nonPrivileged": false,
					"felix": map[string]interface{}{
						"ipinip": map[string]interface{}{
							"enabled": ipinip,
						},
						"bpf": map[string]interface{}{
							"enabled": bpf,
						},
						"bpfKubeProxyIPTablesCleanup": map[string]interface{}{
							"enabled": !kubeProxyEnabled,
						},
					},
					"ipv4": map[string]interface{}{
						"pool":                pool,
						"mode":                modeFunc(),
						"autoDetectionMethod": nil,
					},
				},
				"pspDisabled": isPSPDisabled,
			}
			if detectionMethodFunc() != nil {
				expected["config"].(map[string]interface{})["ipv4"].(map[string]interface{})["autoDetectionMethod"] = *detectionMethodFunc()
			}
			for k, v := range additionalGlobalOptions {
				expected["global"].(map[string]string)[k] = v
			}
			Expect(values).To(Equal(expected))
		},

		Entry("empty network config should properly render calico chart values",
			networkConfigNilFunc, networkConfigNilValuesFunc,
			false, true, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(always) }, func() *string { return nil },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("empty network config should properly render calico chart values even without node cidr",
			networkConfigNilFunc, networkConfigNilValuesFunc,
			false, true, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(always) }, func() *string { return nil },
			func() *string { return nil }, nil),
		Entry("should disable felix ip in ip and set pool mode to never when setting backend to none",
			networkConfigBackendNoneFunc, networkConfigBackendNoneFunc,
			false, true, false, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(never) }, func() *string { return nil },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values",
			networkConfigAllFunc, networkConfigAllFunc,
			true, true, false, defaultMtu, true, false, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with mtu",
			networkConfigAllMTUFunc, networkConfigAllMTUFunc,
			false, true, false, mtuVar, true, false, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with ebpf dataplane enabled and kube-proxy disabled",
			networkConfigAllEBPFDataplaneFunc, networkConfigAllEBPFDataplaneFunc,
			false, false, false, defaultMtu, true, true, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with overlay disabled",
			networkConfigOverlayDisabledFunc, networkConfigOverlayDisabledFunc,
			true, true, false, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(*networkConfigOverlayDisabled.IPv4.Mode) }, func() *string { return networkConfigOverlayDisabled.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR, "overlayEnabled": "false", "snatToUpstreamDNSEnabled": "true"}),
		Entry("should correctly compute all of the calico chart values with overlay disabled, but no node cidr",
			networkConfigOverlayDisabledFunc, networkConfigOverlayDisabledFunc,
			true, true, false, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(*networkConfigOverlayDisabled.IPv4.Mode) }, func() *string { return networkConfigOverlayDisabled.IPv4.AutoDetectionMethod },
			func() *string { return nil }, map[string]string{"overlayEnabled": "false", "snatToUpstreamDNSEnabled": "true"}),
		Entry("should respect deprecated fields in order to keep backwards compatibility",
			networkConfigDeprecatedFunc, networkConfigDeprecatedFunc,
			true, true, false, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(*networkConfigDeprecated.IPIP) }, func() *string { return networkConfigDeprecated.IPAutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
	)

	Describe("#ComputeCalicoChartValues", func() {
		DescribeTable("should correctly compute calico chart values with non-privileged mode enabled",
			func(config func() *calicov1alpha1.NetworkConfig, expectedResult bool) {
				values, err := charts.ComputeCalicoChartValues(network, config(), kubernetesVersion, true, true, false, true, &nodeCIDR)
				Expect(err).To(BeNil())

				actual, err := utils.GetFromValuesMap(values, "config", "nonPrivileged")
				Expect(err).To(BeNil())
				Expect(actual).To(Equal(expectedResult))
			},

			Entry("default", networkConfigAllFunc, true),
			Entry("ebpf dataplane enabled", networkConfigAllEBPFDataplaneFunc, false),
		)

		It("should error on invalid config value", func() {
			_, err := charts.ComputeCalicoChartValues(network, networkConfigInvalid, kubernetesVersion, true, true, false, false, &nodeCIDR)
			Expect(err).To(Equal(fmt.Errorf("error when generating calico config: unsupported value for backend: invalid")))
		})
	})

	Describe("#RenderCalicoChart", func() {
		var (
			ctrl                *gomock.Controller
			mockChartRenderer   *mockchartrenderer.MockInterface
			testManifestContent string
			mkManifest          func(name string) manifest.Manifest
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
			testManifestContent = "test-content"
			mkManifest = func(name string) manifest.Manifest {
				return manifest.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		})
		DescribeTable("Render Calico charts correctly",
			func(nodes *string) {
				mockChartRenderer.EXPECT().Render(calico.CalicoChartPath, calico.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(&chartrenderer.RenderedChart{
					ChartName: "test",
					Manifests: []manifest.Manifest{
						mkManifest(charts.CalicoConfigKey),
					},
				}, nil)

				_, err := charts.RenderCalicoChart(mockChartRenderer, network, networkConfigNil, kubernetesVersion, false, true, false, false, nodes)
				Expect(err).NotTo(HaveOccurred())

			},

			Entry("with node cidr", &nodeCIDR),
			Entry("without node cidr", nil),
		)
	})
})
