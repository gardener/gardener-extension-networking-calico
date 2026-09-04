// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserverendpoints

import (
	"context"

	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// dnsRecordSelector selects the DNSRecords of a shoot's control plane which point to the kube-apiserver. The `ingress`
// record of the nginx-ingress addon carries the same garden role and must not be considered.
var dnsRecordSelector = labels.SelectorFromSet(labels.Set{
	v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane,
}).Add(dnsRecordRoleRequirement())

func dnsRecordRoleRequirement() labels.Requirement {
	requirement, err := labels.NewRequirement(v1beta1constants.LabelRole, selection.In,
		[]string{v1beta1constants.LabelDNSRecordInternal, v1beta1constants.LabelDNSRecordExternal})
	utilruntime.Must(err)

	return *requirement
}

// dnsRecordValues are the kube-apiserver addresses published via DNSRecords, grouped by record type. gardenlet derives
// the type from the address it publishes, see extensionsv1alpha1helper.GetDNSRecordType, so the type states whether the
// kube-apiserver is exposed via an IP address or via a hostname.
type dnsRecordValues struct {
	// addresses are the values of the A and AAAA records, i.e. IP addresses.
	addresses []string
	// hostnames are the values of the CNAME records. A GlobalNetworkSet cannot express them.
	hostnames []string
}

// fromDNSRecords returns the kube-apiserver addresses published by gardenlet via DNSRecord resources in the given
// control plane namespace. They are preferred over a DNS lookup because the DNSRecord is the write side of exactly the
// DNS entry which shoot pods later resolve, and for A/AAAA records its values already are the IP addresses.
func fromDNSRecords(ctx context.Context, c client.Reader, namespace string) (dnsRecordValues, error) {
	dnsRecordList := &extensionsv1alpha1.DNSRecordList{}
	if err := c.List(ctx, dnsRecordList,
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: dnsRecordSelector},
	); err != nil {
		return dnsRecordValues{}, err
	}

	var values dnsRecordValues

	for _, dnsRecord := range dnsRecordList.Items {
		switch dnsRecord.Spec.RecordType {
		case extensionsv1alpha1.DNSRecordTypeA, extensionsv1alpha1.DNSRecordTypeAAAA:
			values.addresses = append(values.addresses, dnsRecord.Spec.Values...)
		case extensionsv1alpha1.DNSRecordTypeCNAME:
			values.hostnames = append(values.hostnames, dnsRecord.Spec.Values...)
		}
	}

	return values, nil
}
