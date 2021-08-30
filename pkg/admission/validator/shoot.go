// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"context"
	"fmt"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewShootValidator returns a new instance of a shoot validator.
func NewShootValidator() extensionswebhook.Validator {
	return &shoot{}
}

type shoot struct {
	client         client.Client
	apiReader      client.Reader
	decoder        runtime.Decoder
	lenientDecoder runtime.Decoder
}

// InjectScheme injects the given scheme into the validator.
func (s *shoot) InjectScheme(scheme *runtime.Scheme) error {
	s.decoder = serializer.NewCodecFactory(scheme, serializer.EnableStrict).UniversalDecoder()
	s.lenientDecoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	return nil
}

// InjectClient injects the given client into the validator.
func (s *shoot) InjectClient(client client.Client) error {
	s.client = client
	return nil
}

// InjectAPIReader injects the given apiReader into the validator.
func (s *shoot) InjectAPIReader(apiReader client.Reader) error {
	s.apiReader = apiReader
	return nil
}

// Validate validates the given shoot object.
func (s *shoot) Validate(ctx context.Context, new, old client.Object) error {
	shoot, ok := new.(*core.Shoot)
	if !ok {
		return fmt.Errorf("wrong object type %T", new)
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

	networkConfig, err := s.decodeNetworkingConfig(shoot.Spec.Networking.ProviderConfig)
	if err != nil {
		return err
	}

	if shoot.Spec.Kubernetes.KubeProxy != nil {
		if shoot.Spec.Kubernetes.KubeProxy.Enabled != nil {
			if !*shoot.Spec.Kubernetes.KubeProxy.Enabled && networkConfig.EbpfDataplane != nil && !networkConfig.EbpfDataplane.Enabled {

				return field.Forbidden(field.NewPath("spec", "kubernetes", "kubeProxy", "enabled"), "Disabling kube-proxy is forbidden in conjunction with calico without running in ebpf dataplane")
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
