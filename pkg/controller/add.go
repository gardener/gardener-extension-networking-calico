// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"

	extensioncontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/network"
	"github.com/gardener/gardener/extensions/pkg/util"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the Calico networking  controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
	// KubeAPIServerEndpoints is the landscape-wide configuration for the kube-apiserver GlobalNetworkSet which is
	// deployed into shoot clusters.
	KubeAPIServerEndpoints *apisconfig.KubeAPIServerEndpointsConfiguration
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	scheme := mgr.GetScheme()
	if err := resourcesv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	chartApplier, err := gardenerkubernetes.NewChartApplierForConfig(mgr.GetConfig())
	if err != nil {
		return fmt.Errorf("could not create ChartApplier: %w", err)
	}

	return network.Add(mgr, network.AddArgs{
		Actuator:          NewActuator(mgr, chartApplier, extensioncontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot), opts.KubeAPIServerEndpoints),
		ControllerOptions: opts.Controller,
		Predicates:        network.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              calico.Type,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
