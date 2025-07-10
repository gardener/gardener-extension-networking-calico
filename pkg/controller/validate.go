// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	apiscalico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	calicovalidation "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/validation"
)

// ValidateNetworkConfig validates the given network configuration.
func ValidateNetworkConfig(networkConfig *calicov1alpha1.NetworkConfig) error {
	internalNetworkConfig := &apiscalico.NetworkConfig{}
	if err := calicov1alpha1.Convert_v1alpha1_NetworkConfig_To_calico_NetworkConfig(networkConfig, internalNetworkConfig, nil); err != nil {
		return fmt.Errorf("could not convert network config: %w", err)
	}

	if errList := calicovalidation.ValidateNetworkConfig(internalNetworkConfig, field.NewPath("spec", "providerConfig")); len(errList) != 0 {
		return fmt.Errorf("invalid network config: %w", errList.ToAggregate())
	}

	return nil
}
