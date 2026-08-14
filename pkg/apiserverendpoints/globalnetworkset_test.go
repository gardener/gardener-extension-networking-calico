// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GlobalNetworkSet", func() {
	var endpoints = func(cidrs []string, source Source) *Endpoints {
		return &Endpoints{CIDRs: cidrs, Source: source}
	}

	Describe("#RenderGlobalNetworkSet", func() {
		It("should fail for nil endpoints", func() {
			_, err := RenderGlobalNetworkSet(nil)

			Expect(err).To(MatchError(ContainSubstring("without CIDRs")))
		})

		It("should fail for an empty list of CIDRs", func() {
			_, err := RenderGlobalNetworkSet(endpoints(nil, SourceDNSRecord))

			Expect(err).To(MatchError(ContainSubstring("without CIDRs")))
		})

		It("should render the expected object", func() {
			data, err := RenderGlobalNetworkSet(endpoints([]string{"34.107.12.34/32"}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`apiVersion: crd.projectcalico.org/v1
kind: GlobalNetworkSet
metadata:
  annotations:
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
			first, err := RenderGlobalNetworkSet(endpoints([]string{"34.107.12.34/32"}, SourceDNSRecord))
			Expect(err).NotTo(HaveOccurred())

			second, err := RenderGlobalNetworkSet(endpoints([]string{"34.107.12.34/32"}, SourceDNSRecord))
			Expect(err).NotTo(HaveOccurred())

			// A timestamp would create a new ManagedResource secret on every reconciliation.
			Expect(string(second)).To(Equal(string(first)))
		})

		It("should render dual-stack nets", func() {
			data, err := RenderGlobalNetworkSet(endpoints([]string{"34.107.12.34/32", "2001:db8::1/128"}, SourceDNSRecord))

			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("- 34.107.12.34/32"))
			Expect(string(data)).To(ContainSubstring("- 2001:db8::1/128"))
		})
	})
})
