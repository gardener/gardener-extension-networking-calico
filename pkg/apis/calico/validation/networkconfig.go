// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
		if err := IsValidMTU(*networkConfig.VethMTU); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("vethMTU"), *networkConfig.VethMTU, fmt.Sprintf("invalid MTU: %q", err)))
		}
	}

	allErrs = append(allErrs, ValidateNetworkConfigAutoscaling(networkConfig.AutoScaling, fldPath.Child("autoScaling"))...)

	if networkConfig.IPIP != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*networkConfig.IPIP) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("ipip"), *networkConfig.IPIP, fmt.Sprintf("unsupported value %q for ipip, supported values are [%q, %q, %q, %q]", *networkConfig.IPIP, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	if networkConfig.IPAutoDetectionMethod != nil {
		allErrs = append(allErrs, ValidateIPAutodetectionMethod(*networkConfig.IPAutoDetectionMethod, fldPath.Child("ipAutoDetectionMethod"))...)
	}

	return allErrs
}

// ValidateNetworkConfigIPAM validates the kube-proxy configuration in the network config.
func ValidateNetworkConfigIPAM(ipam *apiscalico.IPAM, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipam == nil {
		return allErrs
	}

	ipamTypes := sets.New("calico-ipam", "host-local", "kubernetes")

	if ipam.Type != "" && !ipamTypes.Has(ipam.Type) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("type"), ipam.Type, fmt.Sprintf("unsupported value %q for type, supported values are [%q, %q, %q]", ipam.Type, "calico-ipam", "host-local", "kubernetes")))
	}

	if ipam.CIDR != nil && *ipam.CIDR != "" {
		err := validation.IsValidCIDR(fldPath.Child("cidr"), string(*ipam.CIDR))
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("cidr"), *ipam.CIDR, fmt.Sprintf("invalid CIDR: %q", err)))
		}
	}

	return allErrs
}

// ValidateNetworkConfigIPV4 validates the IPv4 configuration in the network config.
func ValidateNetworkConfigIPV4(ipv4 *apiscalico.IPv4, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ipv4 == nil {
		return allErrs
	}

	if ipv4.Pool != nil && *ipv4.Pool != "" && !sets.New(apiscalico.PoolIPIP, apiscalico.PoolVXLan).Has(*ipv4.Pool) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("pool"), *ipv4.Pool, fmt.Sprintf("unsupported value %q for pool, supported values are [%q, %q]", *ipv4.Pool, apiscalico.PoolIPIP, apiscalico.PoolVXLan)))
	}

	if ipv4.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*ipv4.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv4.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q, %q]", *ipv4.Mode, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	if ipv4.AutoDetectionMethod != nil && *ipv4.AutoDetectionMethod != "" {
		allErrs = append(allErrs, ValidateIPAutodetectionMethod(*ipv4.AutoDetectionMethod, fldPath.Child("autoDetectionMethod"))...)
	}

	return allErrs
}

// ValidateNetworkConfigIPV6 validates the IPv6 configuration in the network config.
func ValidateNetworkConfigIPV6(ipv6 *apiscalico.IPv6, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if ipv6 == nil {
		return allErrs
	}

	if ipv6.Pool != nil && *ipv6.Pool != "" && !sets.New(apiscalico.PoolIPIP, apiscalico.PoolVXLan).Has(*ipv6.Pool) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("pool"), *ipv6.Pool, fmt.Sprintf("unsupported value %q for pool, supported values are [%q, %q]", *ipv6.Pool, apiscalico.PoolIPIP, apiscalico.PoolVXLan)))
	}

	if ipv6.Mode != nil && !sets.New(apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off).Has(*ipv6.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), *ipv6.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q, %q]", *ipv6.Mode, apiscalico.Always, apiscalico.Never, apiscalico.CrossSubnet, apiscalico.Off)))
	}

	if ipv6.AutoDetectionMethod != nil && *ipv6.AutoDetectionMethod != "" {
		allErrs = append(allErrs, ValidateIPAutodetectionMethod(*ipv6.AutoDetectionMethod, fldPath.Child("autoDetectionMethod"))...)
	}

	return allErrs
}

func ValidateIPAutodetectionMethod(method string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	validOptions := sets.New("first-found", "can-reach", "interface", "skip-interface", "cidr")
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

	if option == "cidr" {
		if len(parts) != 2 || parts[1] == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, method, "option cidr requires a parameter"))
			return allErrs
		}

		err := validation.IsValidCIDR(nil, parts[1])
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("invalid CIDR: %q", err)))
		}
	}

	if option == "can-reach" || option == "interface" || option == "skip-interface" || option == "cidr" {
		if len(parts) != 2 || parts[1] == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, method, fmt.Sprintf("option %s requires a parameter", option)))
			return allErrs
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
func IsValidMTU(mtu string) error {
	if mtu == "" {
		return fmt.Errorf("MTU cannot be empty")
	}
	mtuInt, err := strconv.Atoi(mtu)
	if err != nil {
		return fmt.Errorf("invalid MTU: %w", err)
	}

	if mtuInt <= 0 {
		return fmt.Errorf("MTU must be a positive integer")
	}

	return nil
}

func ValidateNetworkConfigAutoscaling(autoscaling *apiscalico.AutoScaling, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if autoscaling == nil {
		return allErrs
	}
	if !sets.New(apiscalico.AutoscalingModeClusterProportional, apiscalico.AutoscalingModeVPA, apiscalico.AutoscalingModeStatic).Has(autoscaling.Mode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("mode"), autoscaling.Mode, fmt.Sprintf("unsupported value %q for mode, supported values are [%q, %q, %q]", autoscaling.Mode, apiscalico.AutoscalingModeClusterProportional, apiscalico.AutoscalingModeVPA, apiscalico.AutoscalingModeStatic)))
	}
	if autoscaling.Resources != nil {
		if autoscaling.Resources.Node != nil {
			for name, quantity := range *autoscaling.Resources.Node {
				switch name.String() {
				case "cpu":
					if quantity.IsZero() {
						allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("node").Child("cpu"), quantity.String(), "CPU resource must be specified and cannot be zero"))
					}
				case "memory":
					if quantity.IsZero() {
						allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("node").Child("memory"), quantity.String(), "Memory resource must be specified and cannot be zero"))
					}
				default:
					allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("node").Child(name.String()), quantity.String(), fmt.Sprintf("Unsupported resource: %s", name)))
				}

			}
		}

		if autoscaling.Resources.Typha != nil {
			for name, quantity := range *autoscaling.Resources.Typha {
				switch name.String() {
				case "cpu":
					if quantity.IsZero() {
						allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("typha").Child("cpu"), quantity.String(), "CPU resource must be specified and cannot be zero"))
					}
				case "memory":
					if quantity.IsZero() {
						allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("typha").Child("memory"), quantity.String(), "Memory resource must be specified and cannot be zero"))
					}
				default:
					allErrs = append(allErrs, field.Invalid(fldPath.Child("resources").Child("typha").Child(name.String()), quantity.String(), fmt.Sprintf("Unsupported resource: %s", name)))
				}

			}
		}
	}

	return allErrs
}
