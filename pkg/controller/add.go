// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"

	extensioncontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/network"
	"github.com/gardener/gardener/extensions/pkg/util"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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

	if err := network.Add(mgr, network.AddArgs{
		Actuator:          NewActuator(mgr, chartApplier, extensioncontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot)),
		ControllerOptions: opts.Controller,
		Predicates:        network.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:              calico.Type,
	}); err != nil {
		return err
	}

	// Add watch for ControlPlane resources to trigger Network reconciliation
	// when RouteControllerActive condition changes
	return addControlPlaneWatch(ctx, mgr)
}

// addControlPlaneWatch adds a watch for ControlPlane resources to trigger Network reconciliation.
func addControlPlaneWatch(_ context.Context, mgr manager.Manager) error {
	// Get the Network controller from the manager
	c, err := controller.New(
		"controlplane-to-network-mapper",
		mgr,
		controller.Options{
			Reconciler: reconcile.Func(func(context.Context, reconcile.Request) (reconcile.Result, error) {
				// This is a no-op reconciler, the actual reconciliation is handled by the Network controller
				return reconcile.Result{}, nil
			}),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create controlplane-to-network-mapper controller: %w", err)
	}

	// Watch ControlPlane resources and map them to Network reconciliation requests
	if err := c.Watch(
		source.Kind(
			mgr.GetCache(),
			&extensionsv1alpha1.ControlPlane{},
			handler.TypedEnqueueRequestsFromMapFunc(mapControlPlaneToNetwork),
		),
	); err != nil {
		return fmt.Errorf("failed to add ControlPlane watch: %w", err)
	}

	return nil
}

// mapControlPlaneToNetwork maps ControlPlane changes to Network reconciliation requests.
// When a ControlPlane status changes (e.g., RouteControllerActive condition), this triggers
// a reconciliation of the corresponding Network resource.
func mapControlPlaneToNetwork(ctx context.Context, obj *extensionsv1alpha1.ControlPlane) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: client.ObjectKey{
				Namespace: obj.Namespace,
				Name:      obj.Name,
			},
		},
	}
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
