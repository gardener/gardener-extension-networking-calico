// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
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
		func(networkConfig *calicov1alpha1.NetworkConfig, operatorConfig *bool, expected bool) {
			var operator *apisconfig.KubeAPIServerGlobalNetworkSetConfiguration
			if operatorConfig != nil {
				operator = &apisconfig.KubeAPIServerGlobalNetworkSetConfiguration{Enabled: operatorConfig}
			}

			Expect(Enabled(networkConfig, operator)).To(Equal(expected))
		},
		Entry("no config at all", nil, nil, false),
		Entry("no kubeAPIServerGlobalNetworkSet", &calicov1alpha1.NetworkConfig{}, nil, false),
		Entry("providerConfig enabled", networkConfig(ptr.To(true)), nil, true),
		Entry("providerConfig disabled", networkConfig(ptr.To(false)), nil, false),
		Entry("operator default enabled", nil, ptr.To(true), true),
		Entry("operator default disabled", nil, ptr.To(false), false),
		Entry("providerConfig opts out of the operator default", networkConfig(ptr.To(false)), ptr.To(true), false),
		Entry("providerConfig opts in before the operator default", networkConfig(ptr.To(true)), ptr.To(false), true),
		Entry("providerConfig without enabled falls back to the operator default", networkConfig(nil), ptr.To(true), true),
		Entry("providerConfig without enabled and no operator default", networkConfig(nil), nil, false),
	)

	Describe("#CIDRs", func() {
		const (
			namespace = "shoot--foo--bar"
			hostname  = "abc.elb.eu-west-1.amazonaws.com"
		)

		var (
			ctx = context.Background()

			// Resolution is not attempted for A/AAAA records, so this resolver fails the test if it is used.
			noResolver = &fakeResolver{err: errors.New("the resolver must not be used")}
		)

		BeforeEach(func() {
			// Retry quickly, the tests do not wait for real DNS.
			interval, timeout := ResolveInterval, ResolveTimeout
			ResolveInterval, ResolveTimeout = time.Millisecond, 50*time.Millisecond
			DeferCleanup(func() { ResolveInterval, ResolveTimeout = interval, timeout })
		})

		It("should turn the A and AAAA record values into CIDRs, sorted and deduplicated", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeAAAA, "2001:db8::1", "34.107.12.34"),
			).Build()

			Expect(CIDRs(ctx, c, noResolver, namespace)).To(Equal([]string{"2001:db8::1/128", "34.107.12.34/32"}))
		})

		It("should resolve the hostname of a CNAME record", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
			).Build()
			resolver := &fakeResolver{addresses: map[string][]string{hostname: {"52.1.2.3", "52.1.2.4"}}}

			Expect(CIDRs(ctx, c, resolver, namespace)).To(Equal([]string{"52.1.2.3/32", "52.1.2.4/32"}))
			// The hostname is resolved once, although both DNSRecords carry it.
			Expect(resolver.calls).To(Equal(1))
		})

		It("should combine the addresses of A and CNAME records", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
			).Build()
			resolver := &fakeResolver{addresses: map[string][]string{hostname: {"52.1.2.3", "34.107.12.34"}}}

			Expect(CIDRs(ctx, c, resolver, namespace)).To(Equal([]string{"34.107.12.34/32", "52.1.2.3/32"}))
		})

		It("should retry resolving a hostname", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
			).Build()
			resolver := &fakeResolver{addresses: map[string][]string{hostname: {"52.1.2.3"}}, failFirst: 2}

			Expect(CIDRs(ctx, c, resolver, namespace)).To(Equal([]string{"52.1.2.3/32"}))
			Expect(resolver.calls).To(Equal(3), "two failed attempts and the successful one")
		})

		It("should fail retryably if a hostname cannot be resolved within the timeout", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
			).Build()
			resolver := &fakeResolver{err: errors.New("no such host")}

			_, err := CIDRs(ctx, c, resolver, namespace)

			Expect(err).To(MatchError(ContainSubstring(`could not resolve the kube-apiserver hostname "` + hostname + `"`)))
			Expect(err).To(MatchError(ContainSubstring("no such host")))
			Expect(resolver.calls).To(BeNumerically(">", 1), "the resolution should have been retried")
			// Not a configuration problem: DNS may recover, and gardenlet retries the reconciliation.
			Expect(v1beta1helper.ExtractErrorCodes(err)).To(BeEmpty())
		})

		It("should fail retryably if a hostname resolves to no address", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, hostname),
			).Build()
			resolver := &fakeResolver{addresses: map[string][]string{hostname: {}}}

			_, err := CIDRs(ctx, c, resolver, namespace)

			Expect(err).To(MatchError(ContainSubstring("do not publish an address yet")))
			Expect(v1beta1helper.ExtractErrorCodes(err)).To(BeEmpty())
		})

		It("should fail retryably if no address is published yet", func() {
			_, err := CIDRs(ctx, fake.NewClientBuilder().WithScheme(scheme).Build(), noResolver, namespace)

			Expect(err).To(MatchError(ContainSubstring("do not publish an address yet")))
			// Not a configuration problem: the addresses may still show up during the shoot's creation.
			Expect(v1beta1helper.ExtractErrorCodes(err)).To(BeEmpty())
		})

		It("should fail if a record without values exists only", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA),
			).Build()

			_, err := CIDRs(ctx, c, noResolver, namespace)

			Expect(err).To(MatchError(ContainSubstring("do not publish an address yet")))
		})

		It("should fail if an A record value is not an IP address", func() {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "not-an-ip"),
			).Build()

			_, err := CIDRs(ctx, c, noResolver, namespace)

			Expect(err).To(MatchError(ContainSubstring(`yield "not-an-ip", which is not an IP address`)))
		})

		It("should propagate an error while listing the DNSRecords", func() {
			c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(scheme).Build(), interceptor.Funcs{
				List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
					return errors.New("fake list error")
				},
			})

			_, err := CIDRs(ctx, c, noResolver, namespace)

			Expect(err).To(MatchError(ContainSubstring("could not read the kube-apiserver DNSRecords")))
		})
	})
})

func networkConfig(enabled *bool) *calicov1alpha1.NetworkConfig {
	return &calicov1alpha1.NetworkConfig{KubeAPIServerGlobalNetworkSet: &calicov1alpha1.KubeAPIServerGlobalNetworkSet{Enabled: enabled}}
}

// fakeResolver answers from a fixed table. It fails the first failFirst calls, or every call if err is set.
type fakeResolver struct {
	addresses map[string][]string
	err       error
	failFirst int
	calls     int
}

func (r *fakeResolver) LookupHost(_ context.Context, host string) ([]string, error) {
	r.calls++

	if r.err != nil {
		return nil, r.err
	}
	if r.calls <= r.failFirst {
		return nil, fmt.Errorf("attempt %d fails", r.calls)
	}

	addresses, ok := r.addresses[host]
	if !ok {
		return nil, fmt.Errorf("unexpected host %q", host)
	}
	return addresses, nil
}
