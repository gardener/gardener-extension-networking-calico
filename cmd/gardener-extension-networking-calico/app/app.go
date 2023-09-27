// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package app

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/gardener/gardener/pkg/logger"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	calicoinstall "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/install"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
	calicocmd "github.com/gardener/gardener-extension-networking-calico/pkg/cmd"
	calicocontroller "github.com/gardener/gardener-extension-networking-calico/pkg/controller"
	"github.com/gardener/gardener-extension-networking-calico/pkg/features"
	"github.com/gardener/gardener-extension-networking-calico/pkg/healthcheck"
)

const Name = "gardener-extension-networking-calico"

// NewControllerManagerCommand creates a new command for running a Calico controller.
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		generalOpts = &controllercmd.GeneralOptions{}
		restOpts    = &controllercmd.RESTOptions{}
		mgrOpts     = &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(calico.Name),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}
		// options for the networking-calico controller
		calicoCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}
		reconcileOpts = &controllercmd.ReconcilerOptions{
			IgnoreOperationAnnotation: true,
		}

		// options for the health care controller
		healthCheckCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		heartbeatCtrlOpts = &heartbeatcmd.Options{
			ExtensionName:        calico.Name,
			RenewIntervalSeconds: 30,
			Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}

		configFileOpts = &calicocmd.ConfigOptions{}

		aggOption = controllercmd.NewOptionAggregator(
			generalOpts,
			restOpts,
			mgrOpts,
			calicoCtrlOpts,
			controllercmd.PrefixOption("healthcheck-", healthCheckCtrlOpts),
			controllercmd.PrefixOption("heartbeat-", heartbeatCtrlOpts),
			reconcileOpts,
			configFileOpts,
		)
	)

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", calico.Name),

		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()

			if err := aggOption.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			if err := heartbeatCtrlOpts.Validate(); err != nil {
				return err
			}

			log, err := logger.NewZapLogger(logger.InfoLevel, logger.FormatJSON)
			if err != nil {
				return fmt.Errorf("error instantiating zap logger: %w", err)
			}
			logf.SetLogger(log)

			log.Info("Starting "+Name, "version", version.Get())

			if err := features.FeatureGate.SetFromMap(configFileOpts.Completed().Config.FeatureGates); err != nil {
				return fmt.Errorf("error setting feature gates: %w", err)
			}

			util.ApplyClientConnectionConfigurationToRESTConfig(configFileOpts.Completed().Config.ClientConnection, restOpts.Completed().Config)

			completedMgrOpts := mgrOpts.Completed().Options()
			completedMgrOpts.Client = client.Options{
				Cache: &client.CacheOptions{
					DisableFor: []client.Object{
						&corev1.Secret{},    // applied for ManagedResources
						&corev1.ConfigMap{}, // applied for monitoring config
					},
				},
			}

			mgr, err := manager.New(restOpts.Completed().Config, completedMgrOpts)
			if err != nil {
				return fmt.Errorf("could not instantiate manager: %w", err)
			}

			if err := controller.AddToScheme(mgr.GetScheme()); err != nil {
				return fmt.Errorf("could not update manager scheme: %w", err)
			}

			if err := calicoinstall.AddToScheme(mgr.GetScheme()); err != nil {
				return fmt.Errorf("could not update manager scheme: %w", err)
			}

			reconcileOpts.Completed().Apply(&calicocontroller.DefaultAddOptions.IgnoreOperationAnnotation)
			calicoCtrlOpts.Completed().Apply(&calicocontroller.DefaultAddOptions.Controller)
			configFileOpts.Completed().ApplyHealthCheckConfig(&healthcheck.AddOptions.HealthCheckConfig)
			healthCheckCtrlOpts.Completed().Apply(&healthcheck.AddOptions.Controller)
			heartbeatCtrlOpts.Completed().Apply(&heartbeat.DefaultAddOptions)

			if err := calicocontroller.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add controllers to manager: %w", err)
			}

			if err := healthcheck.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add health check controller to manager: %w", err)
			}

			if err := heartbeat.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add healtbeat controller to manager: %w", err)
			}

			if err := mgr.Start(ctx); err != nil {
				return fmt.Errorf("error running manager: %w", err)
			}
			return nil
		},
	}

	aggOption.AddFlags(cmd.Flags())

	return cmd
}
