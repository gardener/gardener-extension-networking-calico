// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	apiscalico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
	"github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/validation"
)

var _ = Describe("Network validation", func() {
	DescribeTable("#ValidateNetworkConfig",
		func(networkConfig *apiscalico.NetworkConfig, fldPath *field.Path, matcher gomegatypes.GomegaMatcher) {
			Expect(validation.ValidateNetworkConfig(networkConfig, fldPath)).To(matcher)
		},

		Entry("should succeed with empty config", &apiscalico.NetworkConfig{}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with incorrect backend config", &apiscalico.NetworkConfig{Backend: ptr.To(apiscalico.Backend("geneve"))}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.backend")})))),
		Entry("should succeed with valid backend config", &apiscalico.NetworkConfig{Backend: ptr.To(apiscalico.Bird)}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with incorrect ipam config", &apiscalico.NetworkConfig{IPAM: &apiscalico.IPAM{Type: "unknown"}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipam.type")})))),
		Entry("should succeed with valid ipam config", &apiscalico.NetworkConfig{IPAM: &apiscalico.IPAM{Type: "calico-ipam"}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid ipam config for usePodCidr", &apiscalico.NetworkConfig{IPAM: &apiscalico.IPAM{Type: "calico-ipam", CIDR: ptr.To(apiscalico.CIDR("usePodCidr"))}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid ipam config for usePodCidr", &apiscalico.NetworkConfig{IPAM: &apiscalico.IPAM{Type: "calico-ipam", CIDR: ptr.To(apiscalico.CIDR("usePodCIDR"))}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IPv4 pool mode", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{Mode: ptr.To(apiscalico.PoolMode("invalid"))}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.mode")})))),
		Entry("should return error with invalid IPv4 pool", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{Pool: ptr.To(apiscalico.Pool("geneve"))}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.pool")})))),
		Entry("should succeed with valid IPv4 pool mode", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{Mode: ptr.To(apiscalico.Always), Pool: ptr.To(apiscalico.PoolIPIP)}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with pool mode VXLAN and valid mode", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{Pool: ptr.To(apiscalico.PoolVXLan), Mode: ptr.To(apiscalico.Never)}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with pool mode VXLAN and invalid mode", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{Pool: ptr.To(apiscalico.PoolVXLan), Mode: ptr.To(apiscalico.CrossSubnet)}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.mode")})))),
		Entry("should succeed with pool mode VXLAN and valid mode", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{Pool: ptr.To(apiscalico.PoolVXLan), Mode: ptr.To(apiscalico.Never)}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with pool mode VXLAN and invalid mode", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{Pool: ptr.To(apiscalico.PoolVXLan), Mode: ptr.To(apiscalico.CrossSubnet)}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.mode")})))),
		Entry("should return error with invalid IPv6 pool mode", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{Mode: ptr.To(apiscalico.PoolMode("invalid"))}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.mode")})))),
		Entry("should return error with invalid IPv6 pool", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{Pool: ptr.To(apiscalico.Pool("geneve"))}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.pool")})))),
		Entry("should succeed with valid IPv6 pool mode", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{Mode: ptr.To(apiscalico.Always), Pool: ptr.To(apiscalico.PoolIPIP)}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with a valid IPv6 autodetection method", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("interface=cali1234")}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IPv6 autodetection method", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("invalid-method")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.autoDetectionMethod")})))),
		Entry("should return error with invalid vethMTU", &apiscalico.NetworkConfig{VethMTU: ptr.To("-1500")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.vethMTU")})))),
		Entry("should succeed with valid vethMTU", &apiscalico.NetworkConfig{VethMTU: ptr.To("1500")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid autoscaling mode", &apiscalico.NetworkConfig{AutoScaling: &apiscalico.AutoScaling{Mode: "invalid-mode"}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.mode")})))),
		Entry("should succeed with valid autoscaling mode and resources", &apiscalico.NetworkConfig{
			AutoScaling: &apiscalico.AutoScaling{
				Mode: apiscalico.AutoscalingModeStatic,
				Resources: &apiscalico.StaticResources{
					Node:  &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
					Typha: &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
				},
			},
		}, field.NewPath("config"),
			BeEmpty()),
		Entry("should fail with a static autoscaling mode and invalid resource name", &apiscalico.NetworkConfig{
			AutoScaling: &apiscalico.AutoScaling{
				Mode: apiscalico.AutoscalingModeStatic,
				Resources: &apiscalico.StaticResources{
					Node:  &corev1.ResourceList{"invalid": resource.MustParse("128Mi")},
					Typha: &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
				},
			},
		}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.resources.node.invalid")})))),
		Entry("should return error with invalid IPIP mode", &apiscalico.NetworkConfig{IPIP: ptr.To(apiscalico.PoolMode("invalid"))}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipip")})))),
		Entry("should succeed with valid IPIP mode", &apiscalico.NetworkConfig{IPIP: ptr.To(apiscalico.Always)}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid IP autodetection method for ipv4", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{AutoDetectionMethod: ptr.To("interface=cali1234")}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IP autodetection method for ipv4", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{AutoDetectionMethod: ptr.To("invalid-method")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.autoDetectionMethod")})))),
		Entry("should succeed with valid autodetection method for ipv4", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{AutoDetectionMethod: ptr.To("first-found")}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with first-found method and a parameter for ipv4", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{AutoDetectionMethod: ptr.To("first-found=eth0")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.autoDetectionMethod")})))),
		Entry("should return error with kubernetes-internal method and a parameter for ipv4", &apiscalico.NetworkConfig{IPv4: &apiscalico.IPv4{AutoDetectionMethod: ptr.To("kubernetes-internal-ip=1.1.1.1")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv4.autoDetectionMethod")})))),
		Entry("should succeed with valid IP autodetection method", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("interface=cali1234")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid autodetection method for ipv6", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("first-found")}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid autodetection method and a parameter for ipv6", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("interface=cali1234")}}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IP autodetection method for ipv6", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("invalid-method")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.autoDetectionMethod")})))),
		Entry("should return error with first-found method and a parameter for ipv6", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("first-found=eth0")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.autoDetectionMethod")})))),
		Entry("should return error with kubernetes-internal method and a parameter for ipv6", &apiscalico.NetworkConfig{IPv6: &apiscalico.IPv6{AutoDetectionMethod: ptr.To("kubernetes-internal-ip=1.1.1.1")}}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipv6.autoDetectionMethod")})))),
		Entry("should fail with invalid IP autodetection method", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("invalid-method")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should succeed with valid IP autodetection method with interface and regex", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("interface=^cali.*")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should succeed with valid IP autodetection method with a valid cidr", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("cidr=192.168.0.0/16")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with empty IP autodetection method", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should return error with invalid IP autodetection method format", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("invalid-format")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should return error with invalid IP autodetection method regex", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("interface=invalid[regex")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should succeed with valid IP autodetection method with skip-interface", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("skip-interface=cali1234")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IP autodetection method with skip-interface", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("skip-interface=invalid[regex")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should succeed with valid IP autodetection method with can-reach", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("can-reach=1.1.1.1")}, field.NewPath("config"),
			BeEmpty()),
		Entry("should return error with invalid IP autodetection method with can-reach", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("can-reach=invalid[regex")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should return error with invalid IP autodetection method", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("invalid-method")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should return error with invalid IP autodetection method with CIDR", &apiscalico.NetworkConfig{IPAutoDetectionMethod: ptr.To("cidr=290.8.8.8/16")}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.ipAutoDetectionMethod")})))),
		Entry("should check for a positive quantity in the resources", &apiscalico.NetworkConfig{
			AutoScaling: &apiscalico.AutoScaling{
				Mode: apiscalico.AutoscalingModeStatic,
				Resources: &apiscalico.StaticResources{
					Node:  &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
					Typha: &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
				},
			},
		}, field.NewPath("config"),
			BeEmpty(),
		),
		Entry("should return error with negative CPU resource", &apiscalico.NetworkConfig{
			AutoScaling: &apiscalico.AutoScaling{
				Mode: apiscalico.AutoscalingModeStatic,
				Resources: &apiscalico.StaticResources{
					Node:  &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("-100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
					Typha: &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("64Mi")},
				},
			},
		}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.resources.node.cpu"), "Detail": ContainSubstring("must be positive")})),
				PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.resources.node.cpu"), "Detail": ContainSubstring("must be greater than or equal to 0")}))),
		),
		Entry("should return error with negative memory resource", &apiscalico.NetworkConfig{
			AutoScaling: &apiscalico.AutoScaling{
				Mode: apiscalico.AutoscalingModeStatic,
				Resources: &apiscalico.StaticResources{
					Node:  &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
					Typha: &corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("50m"), corev1.ResourceMemory: resource.MustParse("-64Mi")},
				},
			},
		}, field.NewPath("config"),
			ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.resources.typha.memory"), "Detail": ContainSubstring("must be positive")})),
				PointTo(MatchFields(IgnoreExtras, Fields{"Field": Equal("config.autoScaling.resources.typha.memory"), "Detail": ContainSubstring("must be greater than or equal to 0")}))),
		),
	)
})
