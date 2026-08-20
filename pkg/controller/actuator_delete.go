// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
)

// Delete implements Network.Actuator.
func (a *actuator) Delete(ctx context.Context, _ logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	// First delete the monitoring configuration
	if err := applyMonitoringConfig(ctx, a.client, a.chartApplier, network, true); err != nil {
		return err
	}

	// Then delete the managed resources along with their secrets
	for _, name := range managedResourceNames {
		if err := managedresources.Delete(ctx, a.client, network.Namespace, name, true); err != nil {
			return err
		}
	}

	if cluster != nil && !v1beta1helper.ShootNeedsForceDeletion(cluster.Shoot) {
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		for _, name := range managedResourceNames {
			if err := managedresources.WaitUntilDeleted(timeoutCtx, a.client, network.Namespace, name); err != nil {
				return err
			}
		}
	}

	return nil
}

// ForceDelete implements Network.Actuator.
func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	return a.Delete(ctx, log, network, cluster)
}
