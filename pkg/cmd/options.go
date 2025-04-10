// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionshealthcheckcontroller "github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"

	healthcheckcontroller "github.com/gardener/gardener-extension-networking-calico/pkg/healthcheck"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the calico network extension controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(extensionshealthcheckcontroller.ControllerName, healthcheckcontroller.AddToManager),
		controllercmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
