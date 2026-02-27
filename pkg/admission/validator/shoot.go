// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"fmt"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorehelper "github.com/gardener/gardener/pkg/api/core/helper"
	"github.com/gardener/gardener/pkg/apis/core"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	calicovalidation "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/validation"
	"github.com/gardener/gardener-extension-networking-calico/pkg/controller"
)

// NewShootValidator returns a new instance of a shoot validator.
func NewShootValidator(mgr manager.Manager) extensionswebhook.Validator {
	return &shoot{
		client:         mgr.GetClient(),
		apiReader:      mgr.GetAPIReader(),
		decoder:        serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder(),
		lenientDecoder: serializer.NewCodecFactory(mgr.GetScheme()).UniversalDecoder(),
	}
}

type shoot struct {
	client         client.Client
	apiReader      client.Reader
	decoder        runtime.Decoder
	lenientDecoder runtime.Decoder
}

// Validate validates the given shoot object.
func (s *shoot) Validate(ctx context.Context, new, old client.Object) error {
	shoot, ok := new.(*core.Shoot)
	if !ok {
		return fmt.Errorf("wrong object type %T", new)
	}

	// Skip if it's a workerless Shoot
	if gardencorehelper.IsWorkerless(shoot) {
		return nil
	}

	shootV1Beta1 := &gardencorev1beta1.Shoot{}
	err := gardencorev1beta1.Convert_core_Shoot_To_v1beta1_Shoot(shoot, shootV1Beta1, nil)
	if err != nil {
		return err
	}

	if old != nil {
		oldShoot, ok := old.(*core.Shoot)
		if !ok {
			return fmt.Errorf("wrong object type %T for old object", old)
		}
		return s.validateShootUpdate(ctx, oldShoot, shoot)
	}

	return s.validateShootCreation(ctx, shoot)
}

func (s *shoot) validateShoot(_ context.Context, shoot *core.Shoot) error {
	// Network validation
	if shoot.Spec.Networking != nil {
		if errList := calicovalidation.ValidateNetworking(shoot.Spec.Networking, field.NewPath("spec", "networking")); len(errList) != 0 {
			return errList.ToAggregate()
		}

		if shoot.Spec.Networking.ProviderConfig != nil {
			network := &extensionsv1alpha1.Network{}
			network.Spec.ProviderConfig = shoot.Spec.Networking.ProviderConfig
			networkConfig, err := controller.CalicoNetworkConfigFromNetworkResource(network)
			if err != nil {
				return err
			}

			internalNetworkConfig := &calico.NetworkConfig{}
			err = calicov1alpha1.Convert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(networkConfig, internalNetworkConfig, nil)
			if err != nil {
				return err
			}

			if errList := calicovalidation.ValidateNetworkConfig(internalNetworkConfig, field.NewPath("spec", "networking", "providerConfig")); len(errList) != 0 {
				return errList.ToAggregate()
			}
		}
	}

	networkConfig, err := s.decodeNetworkingConfig(shoot.Spec.Networking.ProviderConfig)
	if err != nil {
		return err
	}

	if shoot.Spec.Kubernetes.KubeProxy != nil {
		if shoot.Spec.Kubernetes.KubeProxy.Enabled != nil {
			if !*shoot.Spec.Kubernetes.KubeProxy.Enabled {
				if networkConfig.EbpfDataplane == nil || (networkConfig.EbpfDataplane != nil && !networkConfig.EbpfDataplane.Enabled) {
					return field.Forbidden(field.NewPath("spec", "kubernetes", "kubeProxy", "enabled"), "Disabling kube-proxy is forbidden in conjunction with calico without running in ebpf dataplane")
				}
			}
		}
	}

	return nil
}

func (s *shoot) validateShootUpdate(ctx context.Context, oldShoot, shoot *core.Shoot) error {
	return s.validateShoot(ctx, shoot)
}

func (s *shoot) validateShootCreation(ctx context.Context, shoot *core.Shoot) error {
	return s.validateShoot(ctx, shoot)
}

func (s *shoot) decodeNetworkingConfig(network *runtime.RawExtension) (*calicov1alpha1.NetworkConfig, error) {
	networkConfig := &calicov1alpha1.NetworkConfig{}
	if network != nil && network.Raw != nil {
		if _, _, err := s.decoder.Decode(network.Raw, nil, networkConfig); err != nil {
			return nil, err
		}
	}
	return networkConfig, nil
}
