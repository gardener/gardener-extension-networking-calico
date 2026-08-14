// SPDX-FileCopyrightText: Contributors to the Gardener project
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
		func(addresses []gardencorev1beta1.ShootAdvertisedAddress, expected []string) {
			Expect(fromCluster(newCluster(addresses))).To(Equal(expected))
		},
		Entry("no address", nil, nil),
		Entry("internal and external address",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressExternal, URL: "https://api.foo.example.com"},
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://api.foo.bar.internal.example.com"},
			},
			[]string{"api.foo.example.com", "api.foo.bar.internal.example.com"}),
		Entry("unmanaged address, the only one which can be an IP address",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			},
			[]string{"172.18.255.1"}),
		Entry("service-account-issuer is ignored",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressServiceAccountIssuer, URL: "https://discovery.example.com/issuer"},
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://api.foo.bar.internal.example.com"},
			},
			[]string{"api.foo.bar.internal.example.com"}),
		Entry("wildcard-tls-seed-bound is ignored, KUBERNETES_SERVICE_HOST never points there",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressWildcardTLSSeedBound, URL: "https://api.foo.seed.example.com"},
			},
			nil),
		Entry("port is stripped",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1:6443"},
			},
			[]string{"172.18.255.1"}),
		Entry("IPv6 literal",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://[2001:db8::1]:443"},
			},
			[]string{"2001:db8::1"}),
		Entry("address without a scheme",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "api.foo.example.com"},
			},
			[]string{"api.foo.example.com"}),
		Entry("unparsable address is dropped",
			[]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressInternal, URL: "https://[::1"},
			},
			nil),
	)

	DescribeTable("#toCIDRs",
		func(addresses, expected []string) {
			cidrs, err := toCIDRs(addresses)

			if expected == nil {
				Expect(err).To(MatchError(ErrNoIPAddress))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(cidrs).To(Equal(expected))
		},
		Entry("IPv4", []string{"34.107.12.34"}, []string{"34.107.12.34/32"}),
		Entry("IPv6", []string{"2001:db8::1"}, []string{"2001:db8::1/128"}),
		Entry("dual-stack", []string{"34.107.12.34", "2001:db8::1"}, []string{"2001:db8::1/128", "34.107.12.34/32"}),
		Entry("sorted and deduplicated", []string{"34.107.12.34", "10.0.0.1", "34.107.12.34"},
			[]string{"10.0.0.1/32", "34.107.12.34/32"}),
		Entry("hostnames are skipped", []string{"api.foo.example.com", "34.107.12.34"}, []string{"34.107.12.34/32"}),
		Entry("only hostnames", []string{"api.foo.example.com"}, nil),
		Entry("no addresses", nil, nil),
	)

	DescribeTable("#toCIDRs error message",
		func(addresses []string, expected string) {
			_, err := toCIDRs(addresses)

			Expect(err).To(MatchError(ErrNoIPAddress))
			Expect(err).To(MatchError(ContainSubstring(expected)))
		},
		Entry("no address at all", nil, "the shoot advertises no kube-apiserver address"),
		Entry("names the addresses which were considered", []string{"api.foo.example.com"},
			"none of [api.foo.example.com] is an IP address"),
	)

	Describe("#Collect", func() {
		const namespace = "shoot--foo--bar"

		var ctx = context.Background()

		It("should prefer the DNSRecords over the advertised addresses", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
			).Build()
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			})

			Expect(Collect(ctx, c, namespace, cluster)).To(Equal(&Endpoints{
				CIDRs:  []string{"34.107.12.34/32"},
				Source: SourceDNSRecord,
			}))
		})

		It("should fall back to the advertised addresses if no DNSRecord exists", func() {
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			})

			Expect(Collect(ctx, fake.NewClientBuilder().WithScheme(scheme).Build(), namespace, cluster)).To(Equal(&Endpoints{
				CIDRs:  []string{"172.18.255.1/32"},
				Source: SourceAdvertisedAddresses,
			}))
		})

		It("should fall back if the DNSRecord has no values yet", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA),
			).Build()
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"},
			})

			Expect(Collect(ctx, c, namespace, cluster)).To(Equal(&Endpoints{
				CIDRs:  []string{"172.18.255.1/32"},
				Source: SourceAdvertisedAddresses,
			}))
		})

		It("should propagate an error while listing the DNSRecords", func() {
			c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(scheme).Build(), interceptor.Funcs{
				List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
					return errors.New("fake list error")
				},
			})

			_, err := Collect(ctx, c, namespace, newCluster(nil))

			Expect(err).To(MatchError(ContainSubstring("could not read DNSRecords")))
			Expect(err).NotTo(MatchError(ErrNoIPAddress))
		})

		It("should return ErrUnsupportedHostname if the DNSRecords are CNAMEs", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
			).Build()

			endpoints, err := Collect(ctx, c, namespace, newCluster(nil))

			Expect(err).To(MatchError(ErrUnsupportedHostname))
			Expect(err).NotTo(MatchError(ErrNoIPAddress))
			// The hostname is named once, although both DNSRecords carry it.
			Expect(err).To(MatchError(ContainSubstring("[abc.elb.eu-west-1.amazonaws.com]")))
			Expect(endpoints).To(BeNil())
		})

		It("should prefer the A record if a CNAME record exists as well", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
			).Build()

			Expect(Collect(ctx, c, namespace, newCluster(nil))).To(Equal(&Endpoints{
				CIDRs:  []string{"34.107.12.34/32"},
				Source: SourceDNSRecord,
			}))
		})

		It("should return ErrNoIPAddress, not ErrUnsupportedHostname, if only advertised hostnames exist", func() {
			cluster := newCluster([]gardencorev1beta1.ShootAdvertisedAddress{
				{Name: v1beta1constants.AdvertisedAddressExternal, URL: "https://api.foo.example.com"},
			})

			_, err := Collect(ctx, fake.NewClientBuilder().WithScheme(scheme).Build(), namespace, cluster)

			Expect(err).To(MatchError(ErrNoIPAddress))
			Expect(err).NotTo(MatchError(ErrUnsupportedHostname))
		})
	})
})

func newCluster(addresses []gardencorev1beta1.ShootAdvertisedAddress) *extensionscontroller.Cluster {
	shoot := &gardencorev1beta1.Shoot{}
	shoot.Status.AdvertisedAddresses = addresses

	return &extensionscontroller.Cluster{Shoot: shoot}
}
