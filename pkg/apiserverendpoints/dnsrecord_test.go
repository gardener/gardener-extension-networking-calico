// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"context"
	"errors"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const namespace = "shoot--foo--bar"

var _ = Describe("DNSRecord", func() {
	var ctx = context.Background()

	DescribeTable("#fromDNSRecords",
		func(dnsRecords []client.Object, expected dnsRecordValues) {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dnsRecords...).Build()

			values, err := fromDNSRecords(ctx, c, namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(values).To(Equal(expected))
		},
		Entry("no DNSRecord", nil, dnsRecordValues{}),
		Entry("A record",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")},
			dnsRecordValues{addresses: []string{"34.107.12.34"}}),
		Entry("AAAA record",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeAAAA, "2001:db8::1")},
			dnsRecordValues{addresses: []string{"2001:db8::1"}}),
		Entry("multiple values",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34", "34.107.12.35")},
			dnsRecordValues{addresses: []string{"34.107.12.34", "34.107.12.35"}}),
		Entry("internal and external record are both collected, toCIDRs deduplicates",
			[]client.Object{
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
			},
			dnsRecordValues{addresses: []string{"34.107.12.34", "34.107.12.34"}}),
		Entry("CNAME record is reported as a hostname, not as an address",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com")},
			dnsRecordValues{hostnames: []string{"abc.elb.eu-west-1.amazonaws.com"}}),
		Entry("AAAA record is an address too",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeAAAA, "2001:db8::1")},
			dnsRecordValues{addresses: []string{"2001:db8::1"}}),
		Entry("A and CNAME record are reported separately",
			[]client.Object{
				newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				newDNSRecord(v1beta1constants.LabelDNSRecordExternal, extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
			},
			dnsRecordValues{addresses: []string{"34.107.12.34"}, hostnames: []string{"abc.elb.eu-west-1.amazonaws.com"}}),
		Entry("ingress record of the nginx-ingress addon is ignored",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordIngress, extensionsv1alpha1.DNSRecordTypeA, "1.2.3.4")},
			dnsRecordValues{}),
		Entry("TXT record is ignored",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeTXT, "some-text")},
			dnsRecordValues{}),
		Entry("record without values is ignored",
			[]client.Object{newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA)},
			dnsRecordValues{}),
	)

	It("should propagate an error while listing", func() {
		c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(scheme).Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("fake list error")
			},
		})

		_, err := fromDNSRecords(ctx, c, namespace)

		Expect(err).To(MatchError("fake list error"))
	})

	It("should ignore DNSRecords in other namespaces", func() {
		dnsRecord := newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")
		dnsRecord.Namespace = "shoot--other--cluster"
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dnsRecord).Build()

		Expect(fromDNSRecords(ctx, c, namespace)).To(Equal(dnsRecordValues{}))
	})

	It("should ignore DNSRecords without the controlplane garden role", func() {
		dnsRecord := newDNSRecord(v1beta1constants.LabelDNSRecordInternal, extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")
		delete(dnsRecord.Labels, v1beta1constants.GardenRole)
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dnsRecord).Build()

		Expect(fromDNSRecords(ctx, c, namespace)).To(Equal(dnsRecordValues{}))
	})
})

func newDNSRecord(role string, recordType extensionsv1alpha1.DNSRecordType, values ...string) *extensionsv1alpha1.DNSRecord {
	return &extensionsv1alpha1.DNSRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo-" + role,
			Namespace: namespace,
			Labels: map[string]string{
				v1beta1constants.LabelRole:  role,
				v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane,
			},
		},
		Spec: extensionsv1alpha1.DNSRecordSpec{
			Name:       "api.foo.bar.example.com",
			RecordType: recordType,
			Values:     values,
		},
	}
}
