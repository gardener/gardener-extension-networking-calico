// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

// Package apiserverendpoints determines the IP addresses of a shoot's kube-apiserver endpoint as reachable from within
// the shoot cluster.
package apiserverendpoints

import (
	"context"
	"fmt"
	"net/netip"
	"slices"

	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

// Enabled returns whether the GlobalNetworkSet shall be deployed:
// providerConfig.enabled ?? operatorConfig.enabled ?? false.
func Enabled(networkConfig *calicov1alpha1.NetworkConfig, operatorConfig *apisconfig.KubeAPIServerGlobalNetworkSetConfiguration) bool {
	if networkConfig != nil && networkConfig.KubeAPIServerGlobalNetworkSet != nil && networkConfig.KubeAPIServerGlobalNetworkSet.Enabled != nil {
		return *networkConfig.KubeAPIServerGlobalNetworkSet.Enabled
	}
	if operatorConfig != nil && operatorConfig.Enabled != nil {
		return *operatorConfig.Enabled
	}
	return false
}

// CIDRs returns the IP addresses of the shoot's kube-apiserver endpoint as /32 respectively /128 CIDRs, read from the
// DNSRecords in the given control plane namespace.
//
// It fails rather than returning nothing, see the caller. A kube-apiserver exposed via a hostname is marked as a
// configuration problem, because only disabling the feature can resolve it, while addresses which are not published
// yet may still appear during the shoot's creation.
func CIDRs(ctx context.Context, c client.Reader, namespace string) ([]string, error) {
	dnsRecords, err := fromDNSRecords(ctx, c, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not read the kube-apiserver DNSRecords: %w", err)
	}

	if len(dnsRecords.addresses) == 0 {
		if len(dnsRecords.hostnames) > 0 {
			return nil, v1beta1helper.NewErrorWithCodes(fmt.Errorf("the kube-apiserver is exposed via the hostname(s) "+
				"%v instead of an IP address, which a GlobalNetworkSet cannot hold - unset "+
				"`kubeAPIServerGlobalNetworkSet.enabled` for this shoot or disable it landscape-wide",
				sortAndCompact(dnsRecords.hostnames)), gardencorev1beta1.ErrorConfigurationProblem)
		}

		return nil, fmt.Errorf("the kube-apiserver DNSRecords do not publish an address yet")
	}

	var cidrs []string

	for _, address := range dnsRecords.addresses {
		addr, err := netip.ParseAddr(address)
		if err != nil {
			return nil, fmt.Errorf("the kube-apiserver DNSRecords publish %q, which is not an IP address", address)
		}

		cidrs = append(cidrs, netip.PrefixFrom(addr, addr.BitLen()).String())
	}

	return sortAndCompact(cidrs), nil
}

// sortAndCompact sorts the given values and removes duplicates, in place.
func sortAndCompact(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}
