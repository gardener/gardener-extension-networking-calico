// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"context"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// dnsRecordRoles are the `role` label values of the DNSRecords pointing to the kube-apiserver. The `ingress` record of
// the nginx-ingress addon must not be considered.
var dnsRecordRoles = sets.New(
	v1beta1constants.LabelDNSRecordInternal,
	v1beta1constants.LabelDNSRecordExternal,
)

// fromDNSRecords returns the kube-apiserver endpoints published by gardenlet via DNSRecord resources in the given
// control plane namespace. They are preferred over a DNS lookup because the DNSRecord is the write side of exactly the
// DNS entry which shoot pods later resolve, and for A/AAAA records its values already are the IP addresses. A DNSRecord
// carries no port, hence defaultPort.
func fromDNSRecords(ctx context.Context, c client.Reader, namespace string) ([]endpoint, error) {
	dnsRecordList := &extensionsv1alpha1.DNSRecordList{}
	if err := c.List(ctx, dnsRecordList,
		client.InNamespace(namespace),
		client.MatchingLabels{v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane},
	); err != nil {
		return nil, err
	}

	var endpoints []endpoint

	for _, dnsRecord := range dnsRecordList.Items {
		if !dnsRecordRoles.Has(dnsRecord.Labels[v1beta1constants.LabelRole]) {
			continue
		}

		switch dnsRecord.Spec.RecordType {
		case extensionsv1alpha1.DNSRecordTypeA, extensionsv1alpha1.DNSRecordTypeAAAA, extensionsv1alpha1.DNSRecordTypeCNAME:
			for _, value := range dnsRecord.Spec.Values {
				endpoints = append(endpoints, endpoint{host: value, port: defaultPort})
			}
		}
	}

	return sortAndCompactEndpoints(endpoints), nil
}
