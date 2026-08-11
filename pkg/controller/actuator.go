// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/network"
	gardenerkubernetes "github.com/gardener/gardener/pkg/client/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
)

const (
	// eventRecorderName is the name this extension reports as the source of the events it records.
	eventRecorderName = "networking-calico"

	// managedResourceOrigin identifies this extension as the creator of the managed resources it deploys.
	managedResourceOrigin = "extension-networking-calico"
)

var (
	// StatusTypeMeta is the TypeMeta of Calico Status
	StatusTypeMeta = metav1.TypeMeta{
		APIVersion: calicov1alpha1.SchemeGroupVersion.String(),
		Kind:       "NetworkStatus",
	}

	// managedResourceNames are all managed resources this extension deploys into shoot clusters. They share a
	// lifecycle: deleted on Delete, kept on Migrate.
	managedResourceNames = []string{
		CalicoConfigManagedResourceName,
		KubeAPIServerEndpointsManagedResourceName,
	}
)

type actuator struct {
	restConfig *rest.Config
	client     client.Client
	// apiReader is an uncached reader. It is used for the DNSRecord resources, which are only read once per
	// reconciliation and therefore do not justify an informer.
	apiReader client.Reader
	// recorder surfaces conditions which deliberately do not fail the reconciliation but which operators and shoot
	// owners still need to learn about.
	recorder events.EventRecorder

	chartRendererFactory extensionscontroller.ChartRendererFactory
	chartApplier         gardenerkubernetes.ChartApplier

	// kubeAPIServerEndpointsConfig is the landscape-wide configuration for the kube-apiserver GlobalNetworkSet.
	kubeAPIServerEndpointsConfig *apisconfig.KubeAPIServerEndpointsConfiguration
}

// NewActuator creates a new Actuator that updates the status of the handled Network resources.
func NewActuator(
	mgr manager.Manager,
	chartApplier gardenerkubernetes.ChartApplier,
	chartRendererFactory extensionscontroller.ChartRendererFactory,
	kubeAPIServerEndpointsConfig *apisconfig.KubeAPIServerEndpointsConfiguration,
) network.Actuator {
	return &actuator{
		client:                       mgr.GetClient(),
		apiReader:                    mgr.GetAPIReader(),
		recorder:                     mgr.GetEventRecorder(eventRecorderName),
		restConfig:                   mgr.GetConfig(),
		chartApplier:                 chartApplier,
		chartRendererFactory:         chartRendererFactory,
		kubeAPIServerEndpointsConfig: kubeAPIServerEndpointsConfig,
	}
}
