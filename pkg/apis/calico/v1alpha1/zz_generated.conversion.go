//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	calico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
	v1 "k8s.io/api/core/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*AutoScaling)(nil), (*calico.AutoScaling)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_AutoScaling_To_calico_AutoScaling(a.(*AutoScaling), b.(*calico.AutoScaling), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.AutoScaling)(nil), (*AutoScaling)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_AutoScaling_To_v1alpha1_AutoScaling(a.(*calico.AutoScaling), b.(*AutoScaling), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*EbpfDataplane)(nil), (*calico.EbpfDataplane)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_EbpfDataplane_To_calico_EbpfDataplane(a.(*EbpfDataplane), b.(*calico.EbpfDataplane), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.EbpfDataplane)(nil), (*EbpfDataplane)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_EbpfDataplane_To_v1alpha1_EbpfDataplane(a.(*calico.EbpfDataplane), b.(*EbpfDataplane), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*IPAM)(nil), (*calico.IPAM)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_IPAM_To_calico_IPAM(a.(*IPAM), b.(*calico.IPAM), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.IPAM)(nil), (*IPAM)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_IPAM_To_v1alpha1_IPAM(a.(*calico.IPAM), b.(*IPAM), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*IPv4)(nil), (*calico.IPv4)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_IPv4_To_calico_IPv4(a.(*IPv4), b.(*calico.IPv4), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.IPv4)(nil), (*IPv4)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_IPv4_To_v1alpha1_IPv4(a.(*calico.IPv4), b.(*IPv4), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*IPv6)(nil), (*calico.IPv6)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_IPv6_To_calico_IPv6(a.(*IPv6), b.(*calico.IPv6), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.IPv6)(nil), (*IPv6)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_IPv6_To_v1alpha1_IPv6(a.(*calico.IPv6), b.(*IPv6), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*NetworkConfig)(nil), (*calico.NetworkConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(a.(*NetworkConfig), b.(*calico.NetworkConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.NetworkConfig)(nil), (*NetworkConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(a.(*calico.NetworkConfig), b.(*NetworkConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*NetworkStatus)(nil), (*calico.NetworkStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(a.(*NetworkStatus), b.(*calico.NetworkStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.NetworkStatus)(nil), (*NetworkStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(a.(*calico.NetworkStatus), b.(*NetworkStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Overlay)(nil), (*calico.Overlay)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Overlay_To_calico_Overlay(a.(*Overlay), b.(*calico.Overlay), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.Overlay)(nil), (*Overlay)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_Overlay_To_v1alpha1_Overlay(a.(*calico.Overlay), b.(*Overlay), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*SnatToUpstreamDNS)(nil), (*calico.SnatToUpstreamDNS)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_SnatToUpstreamDNS_To_calico_SnatToUpstreamDNS(a.(*SnatToUpstreamDNS), b.(*calico.SnatToUpstreamDNS), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.SnatToUpstreamDNS)(nil), (*SnatToUpstreamDNS)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_SnatToUpstreamDNS_To_v1alpha1_SnatToUpstreamDNS(a.(*calico.SnatToUpstreamDNS), b.(*SnatToUpstreamDNS), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*StaticResources)(nil), (*calico.StaticResources)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_StaticResources_To_calico_StaticResources(a.(*StaticResources), b.(*calico.StaticResources), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.StaticResources)(nil), (*StaticResources)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_StaticResources_To_v1alpha1_StaticResources(a.(*calico.StaticResources), b.(*StaticResources), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Typha)(nil), (*calico.Typha)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Typha_To_calico_Typha(a.(*Typha), b.(*calico.Typha), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.Typha)(nil), (*Typha)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_Typha_To_v1alpha1_Typha(a.(*calico.Typha), b.(*Typha), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VXLAN)(nil), (*calico.VXLAN)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VXLAN_To_calico_VXLAN(a.(*VXLAN), b.(*calico.VXLAN), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*calico.VXLAN)(nil), (*VXLAN)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_calico_VXLAN_To_v1alpha1_VXLAN(a.(*calico.VXLAN), b.(*VXLAN), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_AutoScaling_To_calico_AutoScaling(in *AutoScaling, out *calico.AutoScaling, s conversion.Scope) error {
	out.Mode = calico.AutoscalingMode(in.Mode)
	out.Resources = (*calico.StaticResources)(unsafe.Pointer(in.Resources))
	return nil
}

// Convert_v1alpha1_AutoScaling_To_calico_AutoScaling is an autogenerated conversion function.
func Convert_v1alpha1_AutoScaling_To_calico_AutoScaling(in *AutoScaling, out *calico.AutoScaling, s conversion.Scope) error {
	return autoConvert_v1alpha1_AutoScaling_To_calico_AutoScaling(in, out, s)
}

func autoConvert_calico_AutoScaling_To_v1alpha1_AutoScaling(in *calico.AutoScaling, out *AutoScaling, s conversion.Scope) error {
	out.Mode = AutoscalingMode(in.Mode)
	out.Resources = (*StaticResources)(unsafe.Pointer(in.Resources))
	return nil
}

// Convert_calico_AutoScaling_To_v1alpha1_AutoScaling is an autogenerated conversion function.
func Convert_calico_AutoScaling_To_v1alpha1_AutoScaling(in *calico.AutoScaling, out *AutoScaling, s conversion.Scope) error {
	return autoConvert_calico_AutoScaling_To_v1alpha1_AutoScaling(in, out, s)
}

func autoConvert_v1alpha1_EbpfDataplane_To_calico_EbpfDataplane(in *EbpfDataplane, out *calico.EbpfDataplane, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_v1alpha1_EbpfDataplane_To_calico_EbpfDataplane is an autogenerated conversion function.
func Convert_v1alpha1_EbpfDataplane_To_calico_EbpfDataplane(in *EbpfDataplane, out *calico.EbpfDataplane, s conversion.Scope) error {
	return autoConvert_v1alpha1_EbpfDataplane_To_calico_EbpfDataplane(in, out, s)
}

func autoConvert_calico_EbpfDataplane_To_v1alpha1_EbpfDataplane(in *calico.EbpfDataplane, out *EbpfDataplane, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_calico_EbpfDataplane_To_v1alpha1_EbpfDataplane is an autogenerated conversion function.
func Convert_calico_EbpfDataplane_To_v1alpha1_EbpfDataplane(in *calico.EbpfDataplane, out *EbpfDataplane, s conversion.Scope) error {
	return autoConvert_calico_EbpfDataplane_To_v1alpha1_EbpfDataplane(in, out, s)
}

func autoConvert_v1alpha1_IPAM_To_calico_IPAM(in *IPAM, out *calico.IPAM, s conversion.Scope) error {
	out.Type = in.Type
	out.CIDR = (*calico.CIDR)(unsafe.Pointer(in.CIDR))
	return nil
}

// Convert_v1alpha1_IPAM_To_calico_IPAM is an autogenerated conversion function.
func Convert_v1alpha1_IPAM_To_calico_IPAM(in *IPAM, out *calico.IPAM, s conversion.Scope) error {
	return autoConvert_v1alpha1_IPAM_To_calico_IPAM(in, out, s)
}

func autoConvert_calico_IPAM_To_v1alpha1_IPAM(in *calico.IPAM, out *IPAM, s conversion.Scope) error {
	out.Type = in.Type
	out.CIDR = (*CIDR)(unsafe.Pointer(in.CIDR))
	return nil
}

// Convert_calico_IPAM_To_v1alpha1_IPAM is an autogenerated conversion function.
func Convert_calico_IPAM_To_v1alpha1_IPAM(in *calico.IPAM, out *IPAM, s conversion.Scope) error {
	return autoConvert_calico_IPAM_To_v1alpha1_IPAM(in, out, s)
}

func autoConvert_v1alpha1_IPv4_To_calico_IPv4(in *IPv4, out *calico.IPv4, s conversion.Scope) error {
	out.Pool = (*calico.Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*calico.PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	return nil
}

// Convert_v1alpha1_IPv4_To_calico_IPv4 is an autogenerated conversion function.
func Convert_v1alpha1_IPv4_To_calico_IPv4(in *IPv4, out *calico.IPv4, s conversion.Scope) error {
	return autoConvert_v1alpha1_IPv4_To_calico_IPv4(in, out, s)
}

func autoConvert_calico_IPv4_To_v1alpha1_IPv4(in *calico.IPv4, out *IPv4, s conversion.Scope) error {
	out.Pool = (*Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	return nil
}

// Convert_calico_IPv4_To_v1alpha1_IPv4 is an autogenerated conversion function.
func Convert_calico_IPv4_To_v1alpha1_IPv4(in *calico.IPv4, out *IPv4, s conversion.Scope) error {
	return autoConvert_calico_IPv4_To_v1alpha1_IPv4(in, out, s)
}

func autoConvert_v1alpha1_IPv6_To_calico_IPv6(in *IPv6, out *calico.IPv6, s conversion.Scope) error {
	out.Pool = (*calico.Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*calico.PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	out.SourceNATEnabled = (*bool)(unsafe.Pointer(in.SourceNATEnabled))
	return nil
}

// Convert_v1alpha1_IPv6_To_calico_IPv6 is an autogenerated conversion function.
func Convert_v1alpha1_IPv6_To_calico_IPv6(in *IPv6, out *calico.IPv6, s conversion.Scope) error {
	return autoConvert_v1alpha1_IPv6_To_calico_IPv6(in, out, s)
}

func autoConvert_calico_IPv6_To_v1alpha1_IPv6(in *calico.IPv6, out *IPv6, s conversion.Scope) error {
	out.Pool = (*Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	out.SourceNATEnabled = (*bool)(unsafe.Pointer(in.SourceNATEnabled))
	return nil
}

// Convert_calico_IPv6_To_v1alpha1_IPv6 is an autogenerated conversion function.
func Convert_calico_IPv6_To_v1alpha1_IPv6(in *calico.IPv6, out *IPv6, s conversion.Scope) error {
	return autoConvert_calico_IPv6_To_v1alpha1_IPv6(in, out, s)
}

func autoConvert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(in *NetworkConfig, out *calico.NetworkConfig, s conversion.Scope) error {
	out.Backend = (*calico.Backend)(unsafe.Pointer(in.Backend))
	out.IPAM = (*calico.IPAM)(unsafe.Pointer(in.IPAM))
	out.IPv4 = (*calico.IPv4)(unsafe.Pointer(in.IPv4))
	out.IPv6 = (*calico.IPv6)(unsafe.Pointer(in.IPv6))
	out.Typha = (*calico.Typha)(unsafe.Pointer(in.Typha))
	out.VethMTU = (*string)(unsafe.Pointer(in.VethMTU))
	out.EbpfDataplane = (*calico.EbpfDataplane)(unsafe.Pointer(in.EbpfDataplane))
	out.Overlay = (*calico.Overlay)(unsafe.Pointer(in.Overlay))
	out.SnatToUpstreamDNS = (*calico.SnatToUpstreamDNS)(unsafe.Pointer(in.SnatToUpstreamDNS))
	out.AutoScaling = (*calico.AutoScaling)(unsafe.Pointer(in.AutoScaling))
	out.VXLAN = (*calico.VXLAN)(unsafe.Pointer(in.VXLAN))
	out.IPIP = (*calico.PoolMode)(unsafe.Pointer(in.IPIP))
	out.IPAutoDetectionMethod = (*string)(unsafe.Pointer(in.IPAutoDetectionMethod))
	out.WireguardEncryption = in.WireguardEncryption
	return nil
}

// Convert_v1alpha1_NetworkConfig_To_calico_NetworkConfig is an autogenerated conversion function.
func Convert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(in *NetworkConfig, out *calico.NetworkConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(in, out, s)
}

func autoConvert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(in *calico.NetworkConfig, out *NetworkConfig, s conversion.Scope) error {
	out.Backend = (*Backend)(unsafe.Pointer(in.Backend))
	out.IPAM = (*IPAM)(unsafe.Pointer(in.IPAM))
	out.IPv4 = (*IPv4)(unsafe.Pointer(in.IPv4))
	out.IPv6 = (*IPv6)(unsafe.Pointer(in.IPv6))
	out.Typha = (*Typha)(unsafe.Pointer(in.Typha))
	out.VethMTU = (*string)(unsafe.Pointer(in.VethMTU))
	out.EbpfDataplane = (*EbpfDataplane)(unsafe.Pointer(in.EbpfDataplane))
	out.Overlay = (*Overlay)(unsafe.Pointer(in.Overlay))
	out.SnatToUpstreamDNS = (*SnatToUpstreamDNS)(unsafe.Pointer(in.SnatToUpstreamDNS))
	out.AutoScaling = (*AutoScaling)(unsafe.Pointer(in.AutoScaling))
	out.VXLAN = (*VXLAN)(unsafe.Pointer(in.VXLAN))
	out.IPIP = (*PoolMode)(unsafe.Pointer(in.IPIP))
	out.IPAutoDetectionMethod = (*string)(unsafe.Pointer(in.IPAutoDetectionMethod))
	out.WireguardEncryption = in.WireguardEncryption
	return nil
}

// Convert_calico_NetworkConfig_To_v1alpha1_NetworkConfig is an autogenerated conversion function.
func Convert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(in *calico.NetworkConfig, out *NetworkConfig, s conversion.Scope) error {
	return autoConvert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(in, out, s)
}

func autoConvert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in *NetworkStatus, out *calico.NetworkStatus, s conversion.Scope) error {
	out.IPFamilies = *(*[]string)(unsafe.Pointer(&in.IPFamilies))
	return nil
}

// Convert_v1alpha1_NetworkStatus_To_calico_NetworkStatus is an autogenerated conversion function.
func Convert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in *NetworkStatus, out *calico.NetworkStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in, out, s)
}

func autoConvert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in *calico.NetworkStatus, out *NetworkStatus, s conversion.Scope) error {
	out.IPFamilies = *(*[]string)(unsafe.Pointer(&in.IPFamilies))
	return nil
}

// Convert_calico_NetworkStatus_To_v1alpha1_NetworkStatus is an autogenerated conversion function.
func Convert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in *calico.NetworkStatus, out *NetworkStatus, s conversion.Scope) error {
	return autoConvert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in, out, s)
}

func autoConvert_v1alpha1_Overlay_To_calico_Overlay(in *Overlay, out *calico.Overlay, s conversion.Scope) error {
	out.Enabled = in.Enabled
	out.CreatePodRoutes = (*bool)(unsafe.Pointer(in.CreatePodRoutes))
	return nil
}

// Convert_v1alpha1_Overlay_To_calico_Overlay is an autogenerated conversion function.
func Convert_v1alpha1_Overlay_To_calico_Overlay(in *Overlay, out *calico.Overlay, s conversion.Scope) error {
	return autoConvert_v1alpha1_Overlay_To_calico_Overlay(in, out, s)
}

func autoConvert_calico_Overlay_To_v1alpha1_Overlay(in *calico.Overlay, out *Overlay, s conversion.Scope) error {
	out.Enabled = in.Enabled
	out.CreatePodRoutes = (*bool)(unsafe.Pointer(in.CreatePodRoutes))
	return nil
}

// Convert_calico_Overlay_To_v1alpha1_Overlay is an autogenerated conversion function.
func Convert_calico_Overlay_To_v1alpha1_Overlay(in *calico.Overlay, out *Overlay, s conversion.Scope) error {
	return autoConvert_calico_Overlay_To_v1alpha1_Overlay(in, out, s)
}

func autoConvert_v1alpha1_SnatToUpstreamDNS_To_calico_SnatToUpstreamDNS(in *SnatToUpstreamDNS, out *calico.SnatToUpstreamDNS, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_v1alpha1_SnatToUpstreamDNS_To_calico_SnatToUpstreamDNS is an autogenerated conversion function.
func Convert_v1alpha1_SnatToUpstreamDNS_To_calico_SnatToUpstreamDNS(in *SnatToUpstreamDNS, out *calico.SnatToUpstreamDNS, s conversion.Scope) error {
	return autoConvert_v1alpha1_SnatToUpstreamDNS_To_calico_SnatToUpstreamDNS(in, out, s)
}

func autoConvert_calico_SnatToUpstreamDNS_To_v1alpha1_SnatToUpstreamDNS(in *calico.SnatToUpstreamDNS, out *SnatToUpstreamDNS, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_calico_SnatToUpstreamDNS_To_v1alpha1_SnatToUpstreamDNS is an autogenerated conversion function.
func Convert_calico_SnatToUpstreamDNS_To_v1alpha1_SnatToUpstreamDNS(in *calico.SnatToUpstreamDNS, out *SnatToUpstreamDNS, s conversion.Scope) error {
	return autoConvert_calico_SnatToUpstreamDNS_To_v1alpha1_SnatToUpstreamDNS(in, out, s)
}

func autoConvert_v1alpha1_StaticResources_To_calico_StaticResources(in *StaticResources, out *calico.StaticResources, s conversion.Scope) error {
	out.Node = (*v1.ResourceList)(unsafe.Pointer(in.Node))
	out.Typha = (*v1.ResourceList)(unsafe.Pointer(in.Typha))
	return nil
}

// Convert_v1alpha1_StaticResources_To_calico_StaticResources is an autogenerated conversion function.
func Convert_v1alpha1_StaticResources_To_calico_StaticResources(in *StaticResources, out *calico.StaticResources, s conversion.Scope) error {
	return autoConvert_v1alpha1_StaticResources_To_calico_StaticResources(in, out, s)
}

func autoConvert_calico_StaticResources_To_v1alpha1_StaticResources(in *calico.StaticResources, out *StaticResources, s conversion.Scope) error {
	out.Node = (*v1.ResourceList)(unsafe.Pointer(in.Node))
	out.Typha = (*v1.ResourceList)(unsafe.Pointer(in.Typha))
	return nil
}

// Convert_calico_StaticResources_To_v1alpha1_StaticResources is an autogenerated conversion function.
func Convert_calico_StaticResources_To_v1alpha1_StaticResources(in *calico.StaticResources, out *StaticResources, s conversion.Scope) error {
	return autoConvert_calico_StaticResources_To_v1alpha1_StaticResources(in, out, s)
}

func autoConvert_v1alpha1_Typha_To_calico_Typha(in *Typha, out *calico.Typha, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_v1alpha1_Typha_To_calico_Typha is an autogenerated conversion function.
func Convert_v1alpha1_Typha_To_calico_Typha(in *Typha, out *calico.Typha, s conversion.Scope) error {
	return autoConvert_v1alpha1_Typha_To_calico_Typha(in, out, s)
}

func autoConvert_calico_Typha_To_v1alpha1_Typha(in *calico.Typha, out *Typha, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_calico_Typha_To_v1alpha1_Typha is an autogenerated conversion function.
func Convert_calico_Typha_To_v1alpha1_Typha(in *calico.Typha, out *Typha, s conversion.Scope) error {
	return autoConvert_calico_Typha_To_v1alpha1_Typha(in, out, s)
}

func autoConvert_v1alpha1_VXLAN_To_calico_VXLAN(in *VXLAN, out *calico.VXLAN, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_v1alpha1_VXLAN_To_calico_VXLAN is an autogenerated conversion function.
func Convert_v1alpha1_VXLAN_To_calico_VXLAN(in *VXLAN, out *calico.VXLAN, s conversion.Scope) error {
	return autoConvert_v1alpha1_VXLAN_To_calico_VXLAN(in, out, s)
}

func autoConvert_calico_VXLAN_To_v1alpha1_VXLAN(in *calico.VXLAN, out *VXLAN, s conversion.Scope) error {
	out.Enabled = in.Enabled
	return nil
}

// Convert_calico_VXLAN_To_v1alpha1_VXLAN is an autogenerated conversion function.
func Convert_calico_VXLAN_To_v1alpha1_VXLAN(in *calico.VXLAN, out *VXLAN, s conversion.Scope) error {
	return autoConvert_calico_VXLAN_To_v1alpha1_VXLAN(in, out, s)
}
