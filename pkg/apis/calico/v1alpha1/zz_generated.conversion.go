//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright (c) SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	calico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
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
	return nil
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
	out.Pool = (*calico.IPv4Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*calico.IPv4PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	return nil
}

// Convert_v1alpha1_IPv4_To_calico_IPv4 is an autogenerated conversion function.
func Convert_v1alpha1_IPv4_To_calico_IPv4(in *IPv4, out *calico.IPv4, s conversion.Scope) error {
	return autoConvert_v1alpha1_IPv4_To_calico_IPv4(in, out, s)
}

func autoConvert_calico_IPv4_To_v1alpha1_IPv4(in *calico.IPv4, out *IPv4, s conversion.Scope) error {
	out.Pool = (*IPv4Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*IPv4PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	return nil
}

// Convert_calico_IPv4_To_v1alpha1_IPv4 is an autogenerated conversion function.
func Convert_calico_IPv4_To_v1alpha1_IPv4(in *calico.IPv4, out *IPv4, s conversion.Scope) error {
	return autoConvert_calico_IPv4_To_v1alpha1_IPv4(in, out, s)
}

func autoConvert_v1alpha1_IPv6_To_calico_IPv6(in *IPv6, out *calico.IPv6, s conversion.Scope) error {
	out.Pool = (*calico.IPv6Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*calico.IPv6PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
	return nil
}

// Convert_v1alpha1_IPv6_To_calico_IPv6 is an autogenerated conversion function.
func Convert_v1alpha1_IPv6_To_calico_IPv6(in *IPv6, out *calico.IPv6, s conversion.Scope) error {
	return autoConvert_v1alpha1_IPv6_To_calico_IPv6(in, out, s)
}

func autoConvert_calico_IPv6_To_v1alpha1_IPv6(in *calico.IPv6, out *IPv6, s conversion.Scope) error {
	out.Pool = (*IPv6Pool)(unsafe.Pointer(in.Pool))
	out.Mode = (*IPv6PoolMode)(unsafe.Pointer(in.Mode))
	out.AutoDetectionMethod = (*string)(unsafe.Pointer(in.AutoDetectionMethod))
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
	out.IPIP = (*calico.IPv4PoolMode)(unsafe.Pointer(in.IPIP))
	out.IPAutoDetectionMethod = (*string)(unsafe.Pointer(in.IPAutoDetectionMethod))
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
	out.IPIP = (*IPv4PoolMode)(unsafe.Pointer(in.IPIP))
	out.IPAutoDetectionMethod = (*string)(unsafe.Pointer(in.IPAutoDetectionMethod))
	return nil
}

// Convert_calico_NetworkConfig_To_v1alpha1_NetworkConfig is an autogenerated conversion function.
func Convert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(in *calico.NetworkConfig, out *NetworkConfig, s conversion.Scope) error {
	return autoConvert_calico_NetworkConfig_To_v1alpha1_NetworkConfig(in, out, s)
}

func autoConvert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in *NetworkStatus, out *calico.NetworkStatus, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_NetworkStatus_To_calico_NetworkStatus is an autogenerated conversion function.
func Convert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in *NetworkStatus, out *calico.NetworkStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_NetworkStatus_To_calico_NetworkStatus(in, out, s)
}

func autoConvert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in *calico.NetworkStatus, out *NetworkStatus, s conversion.Scope) error {
	return nil
}

// Convert_calico_NetworkStatus_To_v1alpha1_NetworkStatus is an autogenerated conversion function.
func Convert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in *calico.NetworkStatus, out *NetworkStatus, s conversion.Scope) error {
	return autoConvert_calico_NetworkStatus_To_v1alpha1_NetworkStatus(in, out, s)
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
