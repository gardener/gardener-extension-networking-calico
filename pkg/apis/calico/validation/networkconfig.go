// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	apiscalico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
)

// ValidateNetworkConfig validates the network config.
func ValidateNetworkConfig(networkConfig *apiscalico.NetworkConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allowedBackendModes := sets.New(apiscalico.Bird, apiscalico.None, apiscalico.VXLan)
	if networkConfig.Backend != nil && !allowedBackendModes.Has(*networkConfig.Backend) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("backend"), *networkConfig.Backend, fmt.Sprintf("unsupported value %q for backend, supported values are [%q, %q, %q]", *networkConfig.Backend, apiscalico.Bird, apiscalico.None, apiscalico.VXLan)))
	}

	allErrs = append(allErrs, ValidateNetworkConfigIPAM(networkConfig.IPAM, fldPath.Child("ipam"))...)

	allErrs = append(allErrs, ValidateNetworkConfigIPV4(networkConfig.IPv4, fldPath.Child("ipv4"))...)

	allErrs = append(allErrs, ValidateNetworkConfigIPV6(networkConfig.IPv6, fldPath.Child("ipv6"))...)

	if networkConfig.VethMTU != nil {
		allErrs = append(allErrs, IsValidMTU(*networkConfig.VethMTU, fldPath.Child("vethMTU"))...)
	}

	allErrs = append(allErrs, ValidateNetworkConfigAutoscaling(networkConfig.AutoScaling, fldPath.Child("autoScaling"))...)

	if networkConfig.IPIP != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*networkConfig.IPIP) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ipip"), *networkConfig.IPIP, fmt.Sprintf("unsupported value %q for ipip, supported values are [%q, %q, %q, %q]", *networkConfig.IPIP, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	if networkConfig.IPAutoDetectionMethod != nil {
		allErrs = append(allErrs, ValidateIPAutoDetectionMethod(*networkConfig.IPAutoDetectionMethod, fldPath.Child("ipAutoDetectionMethod"))...)
	}

	return allErrs
}

// ValidateNetworkConfigIPAM validates the kube-proxy configuration in the network config.
func ValidateNetworkConfigIPAM(ipam *apiscalico.IPAM, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipam == nil {
		return allErrs
	}

	ipamTypes := sets.New(apiscalico.IPAMCalico, apiscalico.IPAMHostLocal)

	if ipam.Type != "" && !ipamTypes.Has(ipam.Type) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("type"), ipam.Type, fmt.Sprintf("unsupported value %q for type, supported values are [%q, %q]", ipam.Type, apiscalico.IPAMCalico, apiscalico.IPAMHostLocal)))
	}

	if ipam.CIDR != nil {
		allErrs = append(allErrs, validation.IsValidCIDR(fldPath.Child("cidr"), string(*ipam.CIDR))...)
	}

	return allErrs
}

// ValidateNetworkConfigIPV4 validates the IPv4 configuration in the network config.
func ValidateNetworkConfigIPV4(ipv4 *apiscalico.IPv4, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipv4 == nil {
		return allErrs
	}

	if ipv4.Pool != nil && !sets.New(apiscalico.PoolIPIP, apiscalico.PoolVXLan).Has(*ipv4.Pool) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("pool"), *ipv4.Pool, fmt.Sprintf("unsupported value %q for pool, supported values are [%q, %q]", *ipv4.Pool, apiscalico.PoolIPIP, apiscalico.PoolVXLan)))
	}

	if ipv4.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*ipv4.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv4.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q, %q]", *ipv4.Mode, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	// Check for unsupported mode with VXLan pool
	// VXLan pool only supports Always and Never modes, so if the mode is set
	// to CrossSubnet or Off, it is invalid.
	if ipv4.Pool != nil && *ipv4.Pool == apiscalico.PoolVXLan && ipv4.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never).Has(*ipv4.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv4.Mode, fmt.Sprintf("unsupported value %q for mode with pool %q, supported values are [%q, %q]", *ipv4.Mode, apiscalico.PoolVXLan, apiscalico.Always, apiscalico.Never)))
	}

	if ipv4.AutoDetectionMethod != nil && *ipv4.AutoDetectionMethod != "" {
		allErrs = append(allErrs, ValidateIPAutoDetectionMethod(*ipv4.AutoDetectionMethod, fldPath.Child("autoDetectionMethod"))...)
	}

	return allErrs
}

// ValidateNetworkConfigIPV6 validates the IPv6 configuration in the network config.
func ValidateNetworkConfigIPV6(ipv6 *apiscalico.IPv6, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipv6 == nil {
		return allErrs
	}

	if ipv6.Pool != nil && !sets.New(apiscalico.PoolIPIP, apiscalico.PoolVXLan).Has(*ipv6.Pool) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("pool"), *ipv6.Pool, fmt.Sprintf("unsupported value %q for pool, supported values are [%q, %q]", *ipv6.Pool, apiscalico.PoolIPIP, apiscalico.PoolVXLan)))
	}

	if ipv6.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*ipv6.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv6.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q, %q]", *ipv6.Mode, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	// Check for unsupported mode with VXLan pool
	// VXLan pool only supports Always and Never modes, so if the mode is set
	// to CrossSubnet or Off, it is invalid.
	if ipv6.Pool != nil && *ipv6.Pool == apiscalico.PoolVXLan && ipv6.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never).Has(*ipv6.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv6.Mode, fmt.Sprintf("unsupported value %q for mode with pool %q, supported values are [%q, %q]", *ipv6.Mode, apiscalico.PoolVXLan, apiscalico.Always, apiscalico.Never)))
	}

	if ipv6.AutoDetectionMethod != nil && *ipv6.AutoDetectionMethod != "" {
		allErrs = append(allErrs, ValidateIPAutoDetectionMethod(*ipv6.AutoDetectionMethod, fldPath.Child("autoDetectionMethod"))...)
	}

	return allErrs
}

func ValidateIPAutoDetectionMethod(method string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	validOptions := sets.New("first-found", "can-reach", "interface", "skip-interface", "cidr", "kubernetes-internal-ip")
	if method == "" {
		allErrs = append(allErrs, field.Invalid(fldPath, method, "method cannot be empty"))
		return allErrs
	}

	parts := strings.SplitN(method, "=", 2)
	option := parts[0]

	if !validOptions.Has(option) {
		allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("invalid option %q, supported options are %v", option, validOptions.UnsortedList())))
		return allErrs
	}

	if (option == "first-found" || option == "kubernetes-internal-ip") && len(parts) != 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("option %s does not take a parameter", option)))
		return allErrs
	}

	if option == "cidr" {
		if len(parts) != 2 || parts[1] == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, method, "option cidr requires a parameter"))
			return allErrs
		}

		allErrs = append(allErrs, validation.IsValidCIDR(fldPath, parts[1])...)
	}

	if option == "can-reach" || option == "interface" || option == "skip-interface" {
		if len(parts) != 2 || parts[1] == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("option %s requires a parameter", option)))
			return allErrs
		}

		if option == "can-reach" {
			// Validate that the parameter is a valid IP address or DNS name
			ipErrs := validation.IsValidIP(fldPath, parts[1])
			dnsErrs := validation.IsDNS1123Subdomain(parts[1])
			if len(ipErrs) > 0 && len(dnsErrs) > 0 {
				allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("parameter for can-reach must be a valid IP address or DNS name: %v, %v", ipErrs, dnsErrs)))
			}
		}

		if option == "interface" || option == "skip-interface" {
			_, err := regexp.Compile(parts[1])
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("invalid regex for option %s: %v", option, err)))
			}
		}
	}

	return allErrs
}

// IsValidMTU checks if the provided MTU is a valid positive integer.
func IsValidMTU(mtu string, fieldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if mtu == "" {
		return append(allErrs, field.Invalid(fieldPath, mtu, "MTU cannot be empty"))
	}
	mtuInt, err := strconv.Atoi(mtu)
	if err != nil {
		return append(allErrs, field.Invalid(fieldPath, mtu, fmt.Sprintf("invalid MTU: %v", err)))
	}
	if mtuInt < 0 {
		return append(allErrs, field.Invalid(fieldPath, mtu, "MTU must be a positive integer"))
	}
	return allErrs
}

// ValidateNetworkConfigAutoscaling validates the autoscaling configuration in the network config.
// It checks if the mode is one of the supported values and validates the resources if provided.
func ValidateNetworkConfigAutoscaling(autoscaling *apiscalico.AutoScaling, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if autoscaling == nil {
		return allErrs
	}
	if !sets.New(apiscalico.AutoscalingModeClusterProportional, apiscalico.AutoscalingModeVPA, apiscalico.AutoscalingModeStatic).Has(autoscaling.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), autoscaling.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q]", autoscaling.Mode, apiscalico.AutoscalingModeClusterProportional, apiscalico.AutoscalingModeVPA, apiscalico.AutoscalingModeStatic)))
	}
	if autoscaling.Resources != nil {

		allErrs = append(allErrs, ValidateResourceList(autoscaling.Resources.Node, "node", fldPath.Child("resources").Child("node"))...)
		allErrs = append(allErrs, ValidateResourceList(autoscaling.Resources.Typha, "typha", fldPath.Child("resources").Child("typha"))...)
	}

	return allErrs
}

// ValidateResourceList validates the resources in the resource list.
// It checks if the CPU and memory resources are specified and not zero, and if any unsupported resources are present.
func ValidateResourceList(resourceList *v1.ResourceList, resourceType string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if resourceList == nil {
		return allErrs
	}
	for name, quantity := range *resourceList {
		switch name.String() {
		case "cpu":
			if quantity.IsZero() {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("cpu"), quantity.String(), "CPU resource must be specified and cannot be zero"))
			} else if quantity.Sign() < 0 {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("cpu"), quantity.String(), "CPU resource must be positive"))
			}
		case "memory":
			if quantity.IsZero() {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("memory"), quantity.String(), "Memory resource must be specified and cannot be zero"))
			} else if quantity.Sign() < 0 {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("memory"), quantity.String(), "Memory resource must be positive"))
			}
		default:
			allErrs = append(allErrs, field.Invalid(fldPath.Child(name.String()), quantity.String(), fmt.Sprintf("Unsupported resource: %s", name)))
		}
	}

	return allErrs
}
