// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

// Package apiserverendpoints determines the IP addresses of the load balancer in front of a shoot's kube-apiserver, as
// published via the shoot's DNSRecords.
package apiserverendpoints

import (
	"context"
	"fmt"
	"net/netip"
	"slices"
	"time"

	"github.com/gardener/gardener/pkg/utils/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

// HostResolver resolves a hostname to its IP addresses. *net.Resolver implements it.
type HostResolver interface {
	LookupHost(ctx context.Context, host string) ([]string, error)
}

var (
	// ResolveInterval is the interval between two attempts to resolve a hostname.
	ResolveInterval = 2 * time.Second
	// ResolveTimeout bounds the attempts to resolve a hostname. A hostname which cannot be resolved within it fails the
	// reconciliation, which is retried by gardenlet.
	ResolveTimeout = 30 * time.Second
)

// Enabled returns whether the GlobalNetworkSet shall be deployed. The shoot's providerConfig takes precedence over the
// operator's landscape-wide default; if neither sets it, the GlobalNetworkSet is not deployed.
func Enabled(networkConfig *calicov1alpha1.NetworkConfig, operatorConfig *apisconfig.KubeAPIServerGlobalNetworkSetConfiguration) bool {
	if networkConfig != nil && networkConfig.KubeAPIServerGlobalNetworkSet != nil && networkConfig.KubeAPIServerGlobalNetworkSet.Enabled != nil {
		return *networkConfig.KubeAPIServerGlobalNetworkSet.Enabled
	}
	if operatorConfig != nil && operatorConfig.Enabled != nil {
		return *operatorConfig.Enabled
	}
	return false
}

// CIDRs returns the IP addresses of the shoot's kube-apiserver endpoint as /32 respectively /128 CIDRs, determined from
// the DNSRecords in the given control plane namespace. The values of A and AAAA records are used as they are, the
// hostnames of CNAME records are resolved.
//
// It fails rather than returning nothing, see the caller. All failures are retryable: addresses which are not
// published yet may still appear during the shoot's creation, and a hostname which cannot be resolved may become
// resolvable.
func CIDRs(ctx context.Context, c client.Reader, resolver HostResolver, namespace string) ([]string, error) {
	dnsRecords, err := fromDNSRecords(ctx, c, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not read the kube-apiserver DNSRecords: %w", err)
	}

	addresses := dnsRecords.addresses

	for _, hostname := range sortAndCompact(dnsRecords.hostnames) {
		resolved, err := resolve(ctx, resolver, hostname)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, resolved...)
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("the kube-apiserver DNSRecords do not publish an address yet")
	}

	var cidrs []string

	for _, address := range addresses {
		addr, err := netip.ParseAddr(address)
		if err != nil {
			return nil, fmt.Errorf("the kube-apiserver DNSRecords yield %q, which is not an IP address", address)
		}

		cidrs = append(cidrs, netip.PrefixFrom(addr, addr.BitLen()).String())
	}

	return sortAndCompact(cidrs), nil
}

// resolve resolves the given hostname, retrying until ResolveTimeout.
func resolve(ctx context.Context, resolver HostResolver, hostname string) ([]string, error) {
	var addresses []string

	if err := retry.UntilTimeout(ctx, ResolveInterval, ResolveTimeout, func(ctx context.Context) (bool, error) {
		var err error
		if addresses, err = resolver.LookupHost(ctx, hostname); err != nil {
			return retry.MinorError(err)
		}
		return retry.Ok()
	}); err != nil {
		return nil, fmt.Errorf("could not resolve the kube-apiserver hostname %q: %w", hostname, err)
	}

	return addresses, nil
}

// sortAndCompact sorts the given values and removes duplicates, in place.
func sortAndCompact(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}
