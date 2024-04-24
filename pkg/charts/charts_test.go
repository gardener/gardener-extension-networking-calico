// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
	releaseutil "helm.sh/helm/v3/pkg/releaseutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gardener/gardener-extension-networking-calico/charts"
	"github.com/gardener/gardener-extension-networking-calico/imagevector"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	chartspkg "github.com/gardener/gardener-extension-networking-calico/pkg/charts"
)

var (
	trueVar    = true
	mtuVar     = "1430"
	defaultMtu = "1440"
)

var _ = Describe("Chart package test", func() {
	var (
		kubernetesVersion                           = "1.28.0"
		podCIDR                                     = calicov1alpha1.CIDR("12.0.0.0/8")
		nodeCIDR                                    = "10.250.0.0/8"
		usePodCidr                                  = calicov1alpha1.CIDR("usePodCidr")
		crossSubnet                                 = calicov1alpha1.CrossSubnet
		always                                      = calicov1alpha1.Always
		never                                       = calicov1alpha1.Never
		invalid             calicov1alpha1.PoolMode = "invalid"
		autodetectionMethod                         = "interface=eth1"
		backendNone                                 = calicov1alpha1.None
		backendVXLan                                = calicov1alpha1.VXLan
		backendBird                                 = calicov1alpha1.Bird
		backendInvalid                              = calicov1alpha1.Backend("invalid")
		poolIPIP                                    = calicov1alpha1.PoolIPIP
		poolVXlan                                   = calicov1alpha1.PoolVXLan

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
		networkConfigWireguard        *calicov1alpha1.NetworkConfig

		networkConfigNilFunc              = func() *calicov1alpha1.NetworkConfig { return networkConfigNil }
		networkConfigNilValuesFunc        = func() *calicov1alpha1.NetworkConfig { return networkConfigNilValues }
		networkConfigBackendNoneFunc      = func() *calicov1alpha1.NetworkConfig { return networkConfigBackendNone }
		networkConfigAllFunc              = func() *calicov1alpha1.NetworkConfig { return networkConfigAll }
		networkConfigAllMTUFunc           = func() *calicov1alpha1.NetworkConfig { return networkConfigAllMTU }
		networkConfigAllEBPFDataplaneFunc = func() *calicov1alpha1.NetworkConfig { return networkConfigAllEBPFDataplane }
		networkConfigDeprecatedFunc       = func() *calicov1alpha1.NetworkConfig { return networkConfigDeprecated }
		networkConfigOverlayDisabledFunc  = func() *calicov1alpha1.NetworkConfig { return networkConfigOverlayDisabled }
		networkConfigWireguardFunc        = func() *calicov1alpha1.NetworkConfig { return networkConfigWireguard }

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
				IPFamilies:  []extensionsv1alpha1.IPFamily{extensionsv1alpha1.IPFamilyIPv4},
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
		networkConfigWireguard = &calicov1alpha1.NetworkConfig{
			Backend: &backendBird,
			IPAM: &calicov1alpha1.IPAM{
				CIDR: &podCIDR,
				Type: "host-local",
			},
			WireguardEncryption: true,
		}
	})

	DescribeTable("#ComputeCalicoChartValues",
		func(config func() *calicov1alpha1.NetworkConfig, configResult func() *calicov1alpha1.NetworkConfig, wantsVPA bool,
			kubeProxyEnabled bool, mtu string, ipinip bool, bpf bool, pool string,
			modeFunc func() string, detectionMethodFunc func() *string, nodesFunc func() *string, additionalGlobalOptions map[string]string) {
			values, err := chartspkg.ComputeCalicoChartValues(network, config(), kubernetesVersion, wantsVPA, kubeProxyEnabled, false, nodesFunc())
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
						"assign_ipv4": true,
						"assign_ipv6": false,
						"type":        configResult().IPAM.Type,
						"subnet":      string(*configResult().IPAM.CIDR),
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
						"enabled":             true,
						"pool":                pool,
						"mode":                modeFunc(),
						"autoDetectionMethod": nil,
						"wireguard":           configResult().WireguardEncryption,
					},
					"ipv6": map[string]interface{}{
						"enabled":             false,
						"pool":                "",
						"mode":                "",
						"autoDetectionMethod": nil,
						"natOutgoing":         configResult().WireguardEncryption,
						"wireguard":           configResult().WireguardEncryption,
					},
				},
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
			false, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(always) }, func() *string { return nil },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("empty network config should properly render calico chart values even without node cidr",
			networkConfigNilFunc, networkConfigNilValuesFunc,
			false, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(always) }, func() *string { return nil },
			func() *string { return nil }, nil),
		Entry("should disable felix ip in ip and set pool mode to never when setting backend to none",
			networkConfigBackendNoneFunc, networkConfigBackendNoneFunc,
			false, true, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(never) }, func() *string { return nil },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values",
			networkConfigAllFunc, networkConfigAllFunc,
			true, true, defaultMtu, true, false, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with mtu",
			networkConfigAllMTUFunc, networkConfigAllMTUFunc,
			false, true, mtuVar, true, false, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with ebpf dataplane enabled and kube-proxy disabled",
			networkConfigAllEBPFDataplaneFunc, networkConfigAllEBPFDataplaneFunc,
			false, false, defaultMtu, true, true, string(poolVXlan),
			func() string { return string(*networkConfigAll.IPv4.Mode) }, func() *string { return networkConfigAll.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with overlay disabled",
			networkConfigOverlayDisabledFunc, networkConfigOverlayDisabledFunc,
			true, true, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(*networkConfigOverlayDisabled.IPv4.Mode) }, func() *string { return networkConfigOverlayDisabled.IPv4.AutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR, "overlayEnabled": "false", "snatToUpstreamDNSEnabled": "true"}),
		Entry("should correctly compute all of the calico chart values with overlay disabled, but no node cidr",
			networkConfigOverlayDisabledFunc, networkConfigOverlayDisabledFunc,
			true, true, defaultMtu, false, false, string(poolIPIP),
			func() string { return string(*networkConfigOverlayDisabled.IPv4.Mode) }, func() *string { return networkConfigOverlayDisabled.IPv4.AutoDetectionMethod },
			func() *string { return nil }, map[string]string{"overlayEnabled": "false", "snatToUpstreamDNSEnabled": "true"}),
		Entry("should respect deprecated fields in order to keep backwards compatibility",
			networkConfigDeprecatedFunc, networkConfigDeprecatedFunc,
			true, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(*networkConfigDeprecated.IPIP) }, func() *string { return networkConfigDeprecated.IPAutoDetectionMethod },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
		Entry("should correctly compute all of the calico chart values with wireguard enabled",
			networkConfigWireguardFunc, networkConfigWireguardFunc,
			false, true, defaultMtu, true, false, string(poolIPIP),
			func() string { return string(always) }, func() *string { return nil },
			func() *string { return &nodeCIDR }, map[string]string{"nodeCIDR": nodeCIDR}),
	)

	Describe("#ComputeCalicoChartValues", func() {
		var podCIDR = "12.0.0.0/8"
		DescribeTable("should correctly compute calico chart values with non-privileged mode enabled",
			func(config func() *calicov1alpha1.NetworkConfig, expectedResult bool) {
				values, err := chartspkg.ComputeCalicoChartValues(network, config(), kubernetesVersion, true, true, true, &nodeCIDR)
				Expect(err).To(BeNil())

				actual, err := utils.GetFromValuesMap(values, "config", "nonPrivileged")
				Expect(err).To(BeNil())
				Expect(actual).To(Equal(expectedResult))
			},

			Entry("default", networkConfigAllFunc, true),
			Entry("ebpf dataplane enabled", networkConfigAllEBPFDataplaneFunc, false),
		)

		It("should error on invalid config value", func() {
			_, err := chartspkg.ComputeCalicoChartValues(network, networkConfigInvalid, kubernetesVersion, true, true, false, &nodeCIDR)
			Expect(err).To(Equal(fmt.Errorf("error when generating calico config: unsupported value for backend: invalid")))
		})

		Context("IPv4", func() {
			BeforeEach(func() {
				network = &extensionsv1alpha1.Network{
					Spec: extensionsv1alpha1.NetworkSpec{
						IPFamilies: []extensionsv1alpha1.IPFamily{
							extensionsv1alpha1.IPFamilyIPv4,
						},
						PodCIDR: podCIDR,
					},
				}
			})
			It("should correctly configure for IPv4 networks", func() {
				values, err := chartspkg.ComputeCalicoChartValues(
					network,
					nil, "", false, false, false, nil,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(values["config"]).To(And(
					HaveKeyWithValue("ipam", Equal(map[string]interface{}{
						"type":        "host-local",
						"subnet":      "usePodCidr",
						"assign_ipv4": true,
						"assign_ipv6": false,
					})),
					HaveKeyWithValue("ipv4", Equal(map[string]interface{}{
						"enabled":             true,
						"pool":                "ipip",
						"mode":                "Always",
						"autoDetectionMethod": nil,
						"wireguard":           false,
					})),
					HaveKeyWithValue("ipv6",
						HaveKeyWithValue("enabled", false),
					),
				))
				Expect(values["global"]).To(
					HaveKeyWithValue("podCIDR", podCIDR),
				)
			})
			It("should use overrides from the config", func() {
				config := &calicov1alpha1.NetworkConfig{
					IPv4: &calicov1alpha1.IPv4{
						Pool:                pointer(calicov1alpha1.PoolVXLan),
						Mode:                pointer(calicov1alpha1.CrossSubnet),
						AutoDetectionMethod: pointer("first-found"),
					},
				}
				values, err := chartspkg.ComputeCalicoChartValues(
					network, config,
					"", false, false, false, nil,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(values["config"]).To(And(
					HaveKeyWithValue("ipam", Equal(map[string]interface{}{
						"type":        "host-local",
						"subnet":      "usePodCidr",
						"assign_ipv4": true,
						"assign_ipv6": false,
					})),
					HaveKeyWithValue("ipv4", Equal(map[string]interface{}{
						"enabled":             true,
						"pool":                "vxlan",
						"mode":                "CrossSubnet",
						"autoDetectionMethod": "first-found",
						"wireguard":           false,
					})),
					HaveKeyWithValue("ipv6",
						HaveKeyWithValue("enabled", false),
					),
				))
				Expect(values["global"]).To(
					HaveKeyWithValue("podCIDR", podCIDR),
				)
			})
			It("should use (deprecated) IPIP overrides", func() {
				config := &calicov1alpha1.NetworkConfig{
					IPIP: pointer(calicov1alpha1.Off),
				}
				values, err := chartspkg.ComputeCalicoChartValues(
					network, config,
					"", false, false, false, nil,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(values["config"]).To(And(
					HaveKeyWithValue("ipam", Equal(map[string]interface{}{
						"type":        "host-local",
						"subnet":      "usePodCidr",
						"assign_ipv4": true,
						"assign_ipv6": false,
					})),
					HaveKeyWithValue("ipv4", Equal(map[string]interface{}{
						"enabled":             true,
						"pool":                "ipip",
						"mode":                "Off",
						"autoDetectionMethod": nil,
						"wireguard":           false,
					})),
					HaveKeyWithValue("ipv6",
						HaveKeyWithValue("enabled", false),
					),
				))
				Expect(values["global"]).To(
					HaveKeyWithValue("podCIDR", podCIDR),
				)
			})
		})
		Context("IPv6", func() {
			BeforeEach(func() {
				network = &extensionsv1alpha1.Network{
					Spec: extensionsv1alpha1.NetworkSpec{
						IPFamilies: []extensionsv1alpha1.IPFamily{
							extensionsv1alpha1.IPFamilyIPv6,
						},
						PodCIDR: podCIDR,
					},
				}
			})
			It("should correctly configure for IPv6 networks", func() {
				values, err := chartspkg.ComputeCalicoChartValues(
					network,
					nil, "", false, false, false, nil,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(values["config"]).To(And(
					HaveKeyWithValue("ipam", Equal(map[string]interface{}{
						"type":        "calico-ipam",
						"subnet":      "usePodCidrIPv6",
						"assign_ipv4": false,
						"assign_ipv6": true,
					})),
					HaveKeyWithValue("ipv4",
						HaveKeyWithValue("enabled", false),
					),
					HaveKeyWithValue("ipv6", Equal(map[string]interface{}{
						"enabled":             true,
						"pool":                "vxlan",
						"mode":                "Never",
						"autoDetectionMethod": nil,
						"natOutgoing":         true,
						"wireguard":           false,
					})),
				))
				Expect(values["global"]).To(
					HaveKeyWithValue("podCIDR", "12.0.0.0/8"),
				)
			})
			It("should use overrides from the config", func() {
				config := &calicov1alpha1.NetworkConfig{
					IPv6: &calicov1alpha1.IPv6{
						Pool:                pointer(calicov1alpha1.PoolVXLan),
						Mode:                pointer(calicov1alpha1.CrossSubnet),
						AutoDetectionMethod: pointer("first-found"),
					},
				}
				values, err := chartspkg.ComputeCalicoChartValues(
					network, config,
					"", false, false, false, nil,
				)
				Expect(err).NotTo(HaveOccurred())

				Expect(values["config"]).To(And(
					HaveKeyWithValue("ipam", Equal(map[string]interface{}{
						"type":        "calico-ipam",
						"subnet":      "usePodCidrIPv6",
						"assign_ipv4": false,
						"assign_ipv6": true,
					})),
					HaveKeyWithValue("ipv4",
						HaveKeyWithValue("enabled", false),
					),
					HaveKeyWithValue("ipv6", Equal(map[string]interface{}{
						"enabled":             true,
						"pool":                "vxlan",
						"mode":                "CrossSubnet",
						"autoDetectionMethod": "first-found",
						"natOutgoing":         true,
						"wireguard":           false,
					})),
				))
				Expect(values["global"]).To(
					HaveKeyWithValue("podCIDR", "12.0.0.0/8"),
				)
			})
		})
	})

	Describe("#RenderCalicoChart", func() {
		var (
			ctrl                *gomock.Controller
			mockChartRenderer   *mockchartrenderer.MockInterface
			testManifestContent string
			mkManifest          func(name string) releaseutil.Manifest
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockChartRenderer = mockchartrenderer.NewMockInterface(ctrl)
			testManifestContent = "test-content"
			mkManifest = func(name string) releaseutil.Manifest {
				return releaseutil.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		})
		DescribeTable("Render Calico charts correctly",
			func(nodes *string) {
				mockChartRenderer.EXPECT().RenderEmbeddedFS(charts.InternalChart, calico.CalicoChartPath, calico.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(&chartrenderer.RenderedChart{
					ChartName: "test",
					Manifests: []releaseutil.Manifest{
						mkManifest(chartspkg.CalicoConfigKey),
					},
				}, nil)

				_, err := chartspkg.RenderCalicoChart(mockChartRenderer, network, networkConfigNil, kubernetesVersion, false, true, false, nodes)
				Expect(err).NotTo(HaveOccurred())

			},

			Entry("with node cidr", &nodeCIDR),
			Entry("without node cidr", nil),
		)
	})
})

func pointer[T any](v T) *T {
	return &v
}
