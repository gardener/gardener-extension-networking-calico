// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package apiserverendpoints determines the IP addresses of a shoot's kube-apiserver endpoint as reachable from within
// the shoot cluster and renders them into a Calico GlobalNetworkSet.
package apiserverendpoints

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"
	"strings"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

var (
	// ErrNoIPAddress is returned if the addresses are not published yet, or if only advertised hostnames are known.
	// Callers are expected to tolerate it, as it usually means that the shoot is still being created.
	ErrNoIPAddress = errors.New("no kube-apiserver IP address could be determined")

	// ErrUnsupportedHostname is returned if gardenlet publishes the kube-apiserver as a hostname rather than as an IP
	// address, i.e. if the load balancer of the seed's istio ingress gateway is exposed by hostname. A GlobalNetworkSet
	// holds CIDRs, so such a landscape cannot be supported without resolving the hostname.
	ErrUnsupportedHostname = errors.New("the kube-apiserver is exposed via a hostname instead of an IP address")
)

// Source describes where the endpoints were obtained from.
type Source string

const (
	SourceDNSRecord           Source = "DNSRecord"
	SourceAdvertisedAddresses Source = "AdvertisedAddresses"
)

// kubeAPIServerAddressNames are the shoot.status.advertisedAddresses entries which pods are sent to: gardener injects
// one of them as KUBERNETES_SERVICE_HOST, see Shoot.ComputeOutOfClusterAPIServerAddress. Notably neither
// `service-account-issuer` nor `wildcard-tls-seed-bound` is ever used for that.
var kubeAPIServerAddressNames = sets.New(
	v1beta1constants.AdvertisedAddressExternal,
	v1beta1constants.AdvertisedAddressInternal,
	v1beta1constants.AdvertisedAddressUnmanaged,
)

// Endpoints are the kube-apiserver endpoints of a shoot as reachable from within the shoot cluster.
type Endpoints struct {
	// CIDRs are the IP addresses of the kube-apiserver as /32 respectively /128 CIDRs.
	CIDRs  []string
	Source Source
}

// Enabled returns whether the GlobalNetworkSet shall be deployed:
// providerConfig.enabled ?? operatorConfig.enabled ?? false.
func Enabled(providerConfig *calicov1alpha1.KubeAPIServerEndpoints, operatorConfig *apisconfig.KubeAPIServerEndpointsConfiguration) bool {
	if providerConfig != nil && providerConfig.Enabled != nil {
		return *providerConfig.Enabled
	}
	if operatorConfig != nil && operatorConfig.Enabled != nil {
		return *operatorConfig.Enabled
	}
	return false
}

// Collect returns the kube-apiserver endpoints of the shoot in the given control plane namespace, preferring the
// DNSRecord resources over the shoot's advertised addresses.
//
// It wraps ErrUnsupportedHostname if the kube-apiserver is published as a hostname, which callers are expected to treat
// as a configuration problem, and ErrNoIPAddress if no address is published (yet), which they are expected to tolerate.
func Collect(ctx context.Context, c client.Reader, namespace string, cluster *extensionscontroller.Cluster) (*Endpoints, error) {
	dnsRecords, err := fromDNSRecords(ctx, c, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not read DNSRecords: %w", err)
	}

	addresses, source := dnsRecords.addresses, SourceDNSRecord

	switch {
	case len(addresses) > 0:
	case len(dnsRecords.hostnames) > 0:
		// The DNSRecord type is derived from the published address, so CNAME records state that the kube-apiserver has
		// no IP address at all - as opposed to one which is not published yet.
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedHostname, sortAndCompact(dnsRecords.hostnames))
	default:
		addresses, source = fromCluster(cluster), SourceAdvertisedAddresses
	}

	cidrs, err := toCIDRs(addresses)
	if err != nil {
		return nil, err
	}

	return &Endpoints{CIDRs: cidrs, Source: source}, nil
}

// fromCluster returns the kube-apiserver addresses from shoot.status.advertisedAddresses. It is only used for shoots
// without a usable DNSRecord, e.g. with unmanaged DNS, where the address of the kube-apiserver service is advertised
// directly and hence can be an IP address.
func fromCluster(cluster *extensionscontroller.Cluster) []string {
	var addresses []string

	for _, advertisedAddress := range cluster.Shoot.Status.AdvertisedAddresses {
		if !kubeAPIServerAddressNames.Has(advertisedAddress.Name) {
			continue
		}
		if host := hostFromURL(advertisedAddress.URL); host != "" {
			addresses = append(addresses, host)
		}
	}

	return addresses
}

// toCIDRs turns the IP addresses among the given addresses into /32 respectively /128 CIDRs, skipping hostnames. It
// reports ErrNoIPAddress naming the addresses which were considered - hostnames only reach it from
// shoot.status.advertisedAddresses, where they mean that the DNSRecords are not written yet.
func toCIDRs(addresses []string) ([]string, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("%w: the shoot advertises no kube-apiserver address", ErrNoIPAddress)
	}

	var cidrs []string

	for _, address := range addresses {
		addr, err := netip.ParseAddr(address)
		if err != nil {
			continue
		}

		cidrs = append(cidrs, netip.PrefixFrom(addr, addr.BitLen()).String())
	}

	if len(cidrs) == 0 {
		return nil, fmt.Errorf("%w: none of %v is an IP address", ErrNoIPAddress, sortAndCompact(addresses))
	}

	return sortAndCompact(cidrs), nil
}

// hostFromURL returns the host of the given URL, tolerating a missing scheme and stripping a port. It returns an empty
// string if the URL has no host.
func hostFromURL(rawURL string) string {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// sortAndCompact sorts the given values and removes duplicates, in place.
func sortAndCompact(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}
