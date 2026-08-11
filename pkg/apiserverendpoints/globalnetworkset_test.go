// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

var _ = Describe("GlobalNetworkSet", func() {
	var endpoints = func(cidrs []string, ports []int32, source Source) *Endpoints {
		return &Endpoints{CIDRs: cidrs, Ports: ports, Source: source}
	}

	Describe("#Name", func() {
		It("should return the default name for a nil config", func() {
			Expect(Name(nil)).To(Equal(calico.KubeAPIServerEndpointsDefaultName))
		})

		It("should return the default name if no name is configured", func() {
			Expect(Name(&calicov1alpha1.KubeAPIServerEndpoints{})).To(Equal(calico.KubeAPIServerEndpointsDefaultName))
		})

		It("should return the configured name", func() {
			Expect(Name(&calicov1alpha1.KubeAPIServerEndpoints{Name: ptr.To("my-set")})).To(Equal("my-set"))
		})

		It("should fall back to the default name for an empty configured name", func() {
			Expect(Name(&calicov1alpha1.KubeAPIServerEndpoints{Name: ptr.To("")})).To(Equal(calico.KubeAPIServerEndpointsDefaultName))
		})
	})

	Describe("#RenderGlobalNetworkSet", func() {
		It("should fail for nil endpoints", func() {
			_, err := RenderGlobalNetworkSet(nil, nil)

			Expect(err).To(MatchError(ContainSubstring("without CIDRs")))
		})

		It("should fail for an empty list of CIDRs", func() {
			_, err := RenderGlobalNetworkSet(nil, endpoints(nil, []int32{443}, SourceDNSRecord))

			Expect(err).To(MatchError(ContainSubstring("without CIDRs")))
		})

		It("should render the expected object", func() {
			data, err := RenderGlobalNetworkSet(nil, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`apiVersion: crd.projectcalico.org/v1
kind: GlobalNetworkSet
metadata:
  annotations:
    networking.gardener.cloud/ports: "443"
    networking.gardener.cloud/source: DNSRecord
  labels:
    networking.gardener.cloud/endpoint: kube-apiserver
  name: gardener-kube-apiserver
spec:
  nets:
  - 34.107.12.34/32
`))
		})

		It("should render the same bytes for the same input", func() {
			first, err := RenderGlobalNetworkSet(nil, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceDNSRecord))
			Expect(err).NotTo(HaveOccurred())

			second, err := RenderGlobalNetworkSet(nil, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceDNSRecord))
			Expect(err).NotTo(HaveOccurred())

			// A timestamp would create a new ManagedResource secret on every reconciliation.
			Expect(string(second)).To(Equal(string(first)))
		})

		It("should render all ports the kube-apiserver is reachable at", func() {
			data, err := RenderGlobalNetworkSet(nil, endpoints([]string{"172.18.255.1/32"}, []int32{443, 6443}, SourceAdvertisedAddresses))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(`networking.gardener.cloud/ports: 443,6443`))
		})

		It("should render the port taken from the advertised address instead of assuming 443", func() {
			data, err := RenderGlobalNetworkSet(nil, endpoints([]string{"172.18.255.1/32"}, []int32{6443}, SourceAdvertisedAddresses))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(SatisfyAll(
				ContainSubstring(`networking.gardener.cloud/ports: "6443"`),
				Not(ContainSubstring(`networking.gardener.cloud/ports: "443"`)),
				ContainSubstring("networking.gardener.cloud/source: AdvertisedAddresses"),
			))
		})

		It("should honour the configured name and merge additional labels", func() {
			data, err := RenderGlobalNetworkSet(&calicov1alpha1.KubeAPIServerEndpoints{
				Name:   ptr.To("my-set"),
				Labels: map[string]string{"foo": "bar"},
			}, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceAdvertisedAddresses))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("name: my-set"))
			Expect(string(data)).To(ContainSubstring("foo: bar"))
			Expect(string(data)).To(ContainSubstring(calico.KubeAPIServerEndpointsLabelKey + ": " + calico.KubeAPIServerEndpointsLabelValue))
			Expect(string(data)).To(ContainSubstring("networking.gardener.cloud/source: AdvertisedAddresses"))
		})

		It("should not allow the contract label to be overridden", func() {
			data, err := RenderGlobalNetworkSet(&calicov1alpha1.KubeAPIServerEndpoints{
				Labels: map[string]string{calico.KubeAPIServerEndpointsLabelKey: "something-else"},
			}, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(calico.KubeAPIServerEndpointsLabelKey + ": " + calico.KubeAPIServerEndpointsLabelValue))
			Expect(string(data)).NotTo(ContainSubstring("something-else"))
		})

		It("should not set any label besides the contract label and the configured ones", func() {
			data, err := RenderGlobalNetworkSet(&calicov1alpha1.KubeAPIServerEndpoints{
				Labels: map[string]string{"foo": "bar"},
			}, endpoints([]string{"34.107.12.34/32"}, []int32{443}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			// Labels of a GlobalNetworkSet are Calico selector input, so extra labels leak into foreign policies.
			Expect(string(data)).NotTo(ContainSubstring("gardener.cloud/role"))
		})

		It("should render dual-stack nets", func() {
			data, err := RenderGlobalNetworkSet(nil, endpoints([]string{"34.107.12.34/32", "2001:db8::1/128"}, []int32{443}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("- 34.107.12.34/32"))
			Expect(string(data)).To(ContainSubstring("- 2001:db8::1/128"))
		})
	})
})
