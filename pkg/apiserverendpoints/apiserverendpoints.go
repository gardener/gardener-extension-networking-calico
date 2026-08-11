// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package apiserverendpoints determines the IP addresses of a shoot's kube-apiserver endpoint as reachable from within
// the shoot cluster and renders them into a Calico GlobalNetworkSet.
package apiserverendpoints

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

// ErrNoIPAddress is returned if no IP address could be determined without performing a DNS lookup.
var ErrNoIPAddress = errors.New("no kube-apiserver IP address could be determined")

// defaultPort is assumed for addresses without a port: the seed's istio ingress gateway serves the kube-apiserver
// on 443.
const defaultPort int32 = 443

// Source describes where the endpoints were obtained from.
type Source string

const (
	SourceDNSRecord           Source = "DNSRecord"
	SourceAdvertisedAddresses Source = "AdvertisedAddresses"
)

// kubeAPIServerAddressNames are the shoot.status.advertisedAddresses entries which advertise the kube-apiserver.
// Notably `service-account-issuer` does not. Only `unmanaged` can hold an IP address, the others are hostnames which
// newEndpoints drops again - they are collected so that ErrNoIPAddress names them.
var kubeAPIServerAddressNames = sets.New(
	v1beta1constants.AdvertisedAddressExternal,
	v1beta1constants.AdvertisedAddressInternal,
	v1beta1constants.AdvertisedAddressUnmanaged,
	v1beta1constants.AdvertisedAddressWildcardTLSSeedBound,
)

// Endpoints are the kube-apiserver endpoints of a shoot as reachable from within the shoot cluster.
type Endpoints struct {
	// CIDRs are the IP addresses of the kube-apiserver as /32 respectively /128 CIDRs.
	CIDRs []string
	// Ports are the ports the kube-apiserver is reachable at.
	Ports []int32
	// Source describes where the endpoints were obtained from.
	Source Source
}

// endpoint is a single kube-apiserver endpoint. The host is either an IP address or a hostname.
type endpoint struct {
	host string
	port int32
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
// DNSRecord resources over the shoot's advertised addresses. It wraps ErrNoIPAddress if only hostnames were found.
func Collect(ctx context.Context, c client.Reader, namespace string, cluster *extensionscontroller.Cluster) (*Endpoints, error) {
	endpoints, err := fromDNSRecords(ctx, c, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not read DNSRecords: %w", err)
	}

	source := SourceDNSRecord
	if len(endpoints) == 0 {
		endpoints, source = fromCluster(cluster), SourceAdvertisedAddresses
	}

	return newEndpoints(endpoints, source)
}

// fromCluster returns the kube-apiserver endpoints from shoot.status.advertisedAddresses, falling back to
// `api.<shoot.spec.dns.domain>`. It is only used for shoots without a usable DNSRecord, e.g. with unmanaged DNS.
func fromCluster(cluster *extensionscontroller.Cluster) []endpoint {
	if cluster == nil || cluster.Shoot == nil {
		return nil
	}

	var endpoints []endpoint

	for _, advertisedAddress := range cluster.Shoot.Status.AdvertisedAddresses {
		if !kubeAPIServerAddressNames.Has(advertisedAddress.Name) {
			continue
		}
		if e, ok := endpointFromURL(advertisedAddress.URL); ok {
			endpoints = append(endpoints, e)
		}
	}

	// `api.<domain>` is a hostname, so it never contributes an address. It only makes ErrNoIPAddress name the domain
	// the kube-apiserver is expected at.
	if len(endpoints) == 0 && cluster.Shoot.Spec.DNS != nil && cluster.Shoot.Spec.DNS.Domain != nil && *cluster.Shoot.Spec.DNS.Domain != "" {
		endpoints = append(endpoints, endpoint{host: v1beta1helper.GetAPIServerDomain(*cluster.Shoot.Spec.DNS.Domain), port: defaultPort})
	}

	return sortAndCompactEndpoints(endpoints)
}

// newEndpoints turns the given endpoints into /32 respectively /128 CIDRs and the set of ports they are reachable at.
// Hostnames are skipped, resolving them is not implemented.
func newEndpoints(endpoints []endpoint, source Source) (*Endpoints, error) {
	result := &Endpoints{Source: source}

	for _, e := range endpoints {
		addr, err := netip.ParseAddr(e.host)
		if err != nil {
			continue
		}

		result.CIDRs = append(result.CIDRs, netip.PrefixFrom(addr, addr.BitLen()).String())
		result.Ports = append(result.Ports, e.port)
	}

	if len(result.CIDRs) == 0 {
		return nil, fmt.Errorf("%w from %v", ErrNoIPAddress, hosts(endpoints))
	}

	result.CIDRs = sortAndCompact(result.CIDRs)
	result.Ports = sortAndCompact(result.Ports)

	return result, nil
}

// endpointFromURL extracts the host and port from the given URL. A missing scheme is tolerated, a missing port defaults
// to defaultPort, and URLs without a host or with an invalid port are rejected.
func endpointFromURL(rawURL string) (endpoint, bool) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return endpoint{}, false
	}

	host := u.Hostname()
	if host == "" {
		return endpoint{}, false
	}

	port := defaultPort
	if rawPort := u.Port(); rawPort != "" {
		parsedPort, err := strconv.ParseInt(rawPort, 10, 32)
		if err != nil || parsedPort < 1 || parsedPort > 65535 {
			return endpoint{}, false
		}
		port = int32(parsedPort)
	}

	return endpoint{host: host, port: port}, true
}

func hosts(endpoints []endpoint) []string {
	result := make([]string, 0, len(endpoints))
	for _, e := range endpoints {
		result = append(result, e.host)
	}

	return result
}

// sortAndCompactEndpoints sorts by host and port and removes duplicates, in place.
func sortAndCompactEndpoints(endpoints []endpoint) []endpoint {
	slices.SortFunc(endpoints, func(a, b endpoint) int {
		return cmp.Or(cmp.Compare(a.host, b.host), cmp.Compare(a.port, b.port))
	})

	return slices.Compact(endpoints)
}

// sortAndCompact sorts the given values and removes duplicates, in place.
func sortAndCompact[S ~[]E, E cmp.Ordered](values S) S {
	slices.Sort(values)
	return slices.Compact(values)
}
