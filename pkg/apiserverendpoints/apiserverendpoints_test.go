// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"context"
	"errors"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

var _ = Describe("APIServerEndpoints", func() {
	DescribeTable("#Enabled",
		func(providerConfig *calicov1alpha1.KubeAPIServerEndpoints, operatorConfig *bool, expected bool) {
			var operator *apisconfig.KubeAPIServerEndpointsConfiguration
			if operatorConfig != nil {
				operator = &apisconfig.KubeAPIServerEndpointsConfiguration{Enabled: operatorConfig}
			}

			Expect(Enabled(providerConfig, operator)).To(Equal(expected))
		},
		Entry("both unset", nil, nil, false),
		Entry("providerConfig enabled", &calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(true)}, nil, true),
		Entry("providerConfig disabled", &calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(false)}, nil, false),
		Entry("operator default enabled", nil, ptr.To(true), true),
		Entry("operator default disabled", nil, ptr.To(false), false),
		Entry("providerConfig opts out of the operator default",
			&calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(false)}, ptr.To(true), false),
		Entry("providerConfig opts in before the operator default",
			&calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(true)}, ptr.To(false), true),
		Entry("providerConfig without enabled falls back to the operator default",
			&calicov1alpha1.KubeAPIServerEndpoints{}, ptr.To(true), true),
		Entry("providerConfig without enabled and no operator default",
			&calicov1alpha1.KubeAPIServerEndpoints{}, nil, false),
	)

	DescribeTable("#fromCluster",
		func(cluster *extensionscontroller.Cluster, expected []endpoint) {
			Expect(fromCluster(cluster)).To(Equal(expected))
		},
		Entry("nil cluster", nil, nil),
		Entry("cluster without shoot", &extensionscontroller.Cluster{}, nil),
		Entry("no address and no domain", newCluster(nil, nil), nil),
		Entry("internal and external address",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressExternal, URL: "https://api.foo.example.com"},
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://api.foo.bar.internal.example.com"},
			}, nil),
			[]endpoint{{host: "api.foo.bar.internal.example.com", port: 443}, {host: "api.foo.example.com", port: 443}}),
		Entry("service-account-issuer is ignored",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressServiceAccountIssuer, URL: "https://discovery.example.com/issuer"},
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://api.foo.bar.internal.example.com"},
			}, nil),
			[]endpoint{{host: "api.foo.bar.internal.example.com", port: 443}}),
		Entry("port is preserved",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1:6443"},
			}, nil),
			[]endpoint{{host: "172.18.255.1", port: 6443}}),
		Entry("IPv6 literal",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://[2001:db8::1]:443"},
			}, nil),
			[]endpoint{{host: "2001:db8::1", port: 443}}),
		Entry("duplicates are removed",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressExternal, URL: "https://api.foo.example.com"},
				{Name: v1beta1constants.AdvertisedAddressWildcardTLSSeedBound, URL: "https://api.foo.example.com"},
			}, nil),
			[]endpoint{{host: "api.foo.example.com", port: 443}}),
		Entry("the same host on different ports is kept",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressExternal, URL: "https://172.18.255.1"},
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1:6443"},
			}, nil),
			[]endpoint{{host: "172.18.255.1", port: 443}, {host: "172.18.255.1", port: 6443}}),
		Entry("address without a scheme",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "api.foo.example.com"},
			}, nil),
			[]endpoint{{host: "api.foo.example.com", port: 443}}),
		Entry("unparsable address is dropped",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://[::1"},
			}, nil),
			nil),
		Entry("address with an out-of-range port is dropped",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1:99999"},
			}, nil),
			nil),
		Entry("fall back to api.<domain>", newCluster(nil, ptr.To("foo.bar.example.com")),
			[]endpoint{{host: "api.foo.bar.example.com", port: 443}}),
		Entry("no fall back if an address exists",
			newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://api.internal.example.com"},
			}, ptr.To("foo.bar.example.com")),
			[]endpoint{{host: "api.internal.example.com", port: 443}}),
	)

	DescribeTable("#endpointFromURL",
		func(rawURL string, expected endpoint, expectedOK bool) {
			result, ok := endpointFromURL(rawURL)

			Expect(ok).To(Equal(expectedOK))
			Expect(result).To(Equal(expected))
		},
		Entry("host without port", "https://api.foo.example.com", endpoint{host: "api.foo.example.com", port: 443}, true),
		Entry("host with port", "https://api.foo.example.com:6443", endpoint{host: "api.foo.example.com", port: 6443}, true),
		Entry("IPv4 with port", "https://172.18.255.1:6443", endpoint{host: "172.18.255.1", port: 6443}, true),
		Entry("IPv6 with port", "https://[2001:db8::1]:8443", endpoint{host: "2001:db8::1", port: 8443}, true),
		Entry("without scheme", "172.18.255.1:6443", endpoint{host: "172.18.255.1", port: 6443}, true),
		Entry("unparsable URL", "https://[::1", endpoint{}, false),
		Entry("empty URL", "", endpoint{}, false),
		Entry("port zero", "https://172.18.255.1:0", endpoint{}, false),
		Entry("port out of range", "https://172.18.255.1:65536", endpoint{}, false),
		Entry("port exceeding int32", "https://172.18.255.1:99999999999", endpoint{}, false),
	)

	DescribeTable("#newEndpoints",
		func(endpoints []endpoint, expected *Endpoints, expectErrNoIPAddress bool) {
			result, err := newEndpoints(endpoints, SourceDNSRecord)

			if expectErrNoIPAddress {
				Expect(err).To(MatchError(ErrNoIPAddress))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(result).To(Equal(expected))
		},
		Entry("IPv4", []endpoint{{host: "34.107.12.34", port: 443}},
			&Endpoints{CIDRs: []string{"34.107.12.34/32"}, Ports: []int32{443}, Source: SourceDNSRecord}, false),
		Entry("IPv6", []endpoint{{host: "2001:db8::1", port: 443}},
			&Endpoints{CIDRs: []string{"2001:db8::1/128"}, Ports: []int32{443}, Source: SourceDNSRecord}, false),
		Entry("dual-stack", []endpoint{{host: "34.107.12.34", port: 443}, {host: "2001:db8::1", port: 443}},
			&Endpoints{CIDRs: []string{"2001:db8::1/128", "34.107.12.34/32"}, Ports: []int32{443}, Source: SourceDNSRecord}, false),
		Entry("sorted and deduplicated",
			[]endpoint{{host: "34.107.12.34", port: 443}, {host: "10.0.0.1", port: 443}, {host: "34.107.12.34", port: 443}},
			&Endpoints{CIDRs: []string{"10.0.0.1/32", "34.107.12.34/32"}, Ports: []int32{443}, Source: SourceDNSRecord}, false),
		Entry("multiple ports are collected",
			[]endpoint{{host: "34.107.12.34", port: 6443}, {host: "10.0.0.1", port: 443}},
			&Endpoints{CIDRs: []string{"10.0.0.1/32", "34.107.12.34/32"}, Ports: []int32{443, 6443}, Source: SourceDNSRecord}, false),
		Entry("hostnames are skipped",
			[]endpoint{{host: "api.foo.example.com", port: 443}, {host: "34.107.12.34", port: 6443}},
			&Endpoints{CIDRs: []string{"34.107.12.34/32"}, Ports: []int32{6443}, Source: SourceDNSRecord}, false),
		Entry("only hostnames", []endpoint{{host: "api.foo.example.com", port: 443}}, nil, true),
		Entry("no endpoints", nil, nil, true),
	)

	It("should mention the rejected hosts in the ErrNoIPAddress error", func() {
		_, err := newEndpoints([]endpoint{{host: "api.foo.example.com", port: 443}}, SourceDNSRecord)

		Expect(err).To(MatchError(ErrNoIPAddress))
		Expect(err).To(MatchError(ContainSubstring("api.foo.example.com")))
	})

	Describe("#Collect", func() {
		const namespace = "shoot--foo--bar"

		var ctx = context.Background()

		It("should prefer the DNSRecords over the advertised addresses", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
			).Build()
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			}, nil)

			endpoints, err := Collect(ctx, c, namespace, cluster)

			Expect(err).NotTo(HaveOccurred())
			Expect(endpoints).To(Equal(&Endpoints{
				CIDRs:  []string{"34.107.12.34/32"},
				Ports:  []int32{443},
				Source: SourceDNSRecord,
			}))
		})

		It("should fall back to the advertised addresses if no DNSRecord exists", func() {
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1:6443"},
			}, nil)

			endpoints, err := Collect(ctx, fake.NewClientBuilder().WithScheme(scheme).Build(), namespace, cluster)

			Expect(err).NotTo(HaveOccurred())
			Expect(endpoints).To(Equal(&Endpoints{
				CIDRs:  []string{"172.18.255.1/32"},
				Ports:  []int32{6443},
				Source: SourceAdvertisedAddresses,
			}))
		})

		It("should propagate an error while listing the DNSRecords", func() {
			c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(scheme).Build(), interceptor.Funcs{
				List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
					return errors.New("fake list error")
				},
			})

			_, err := Collect(ctx, c, namespace, newCluster(nil, nil))

			Expect(err).To(MatchError(ContainSubstring("could not read DNSRecords")))
			Expect(err).NotTo(MatchError(ErrNoIPAddress))
		})

		It("should fall back if the DNSRecord has no values yet", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA),
			).Build()
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			}, nil)

			endpoints, err := Collect(ctx, c, namespace, cluster)

			Expect(err).NotTo(HaveOccurred())
			Expect(endpoints).To(Equal(&Endpoints{
				CIDRs:  []string{"172.18.255.1/32"},
				Ports:  []int32{443},
				Source: SourceAdvertisedAddresses,
			}))
		})

		It("should return ErrNoIPAddress if only hostnames are available", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
			).Build()

			endpoints, err := Collect(ctx, c, namespace, newCluster(nil, nil))

			Expect(err).To(MatchError(ErrNoIPAddress))
			Expect(err).To(MatchError(ContainSubstring("abc.elb.eu-west-1.amazonaws.com")))
			Expect(endpoints).To(BeNil())
		})

		It("should name the expected kube-apiserver domain if the shoot advertises no address at all", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).Build()

			_, err := Collect(ctx, c, namespace, newCluster(nil, ptr.To("foo.bar.example.com")))

			Expect(err).To(MatchError(ErrNoIPAddress))
			Expect(err).To(MatchError(ContainSubstring("api.foo.bar.example.com")))
		})
	})
})

func newCluster(addresses []gardencorev1beta1.ShootAdvertisedAddress, domain *string) *extensionscontroller.Cluster {
	shoot := &gardencorev1beta1.Shoot{}
	shoot.Status.AdvertisedAddresses = addresses
	if domain != nil {
		shoot.Spec.DNS = &gardencorev1beta1.DNS{Domain: domain}
	}

	return &extensionscontroller.Cluster{Shoot: shoot}
}
