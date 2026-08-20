// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/apiserverendpoints"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

const (
	// KubeAPIServerEndpointsManagedResourceName is the name of the managed resource containing the kube-apiserver
	// GlobalNetworkSet. It is separate from CalicoConfigManagedResourceName so that it can be left untouched if the
	// addresses cannot be determined - an object missing from the calico chart would be deleted from the shoot instead.
	KubeAPIServerEndpointsManagedResourceName = "extension-networking-calico-apiserver-endpoints"

	kubeAPIServerEndpointsDataKey = "globalnetworkset.yaml"

	// EventKubeAPIServerEndpointsOutdated is the event reason used if a previously published GlobalNetworkSet is kept.
	EventKubeAPIServerEndpointsOutdated = "KubeAPIServerEndpointsOutdated"
	// EventKubeAPIServerEndpointsMissing is the event reason used if no GlobalNetworkSet was published at all.
	EventKubeAPIServerEndpointsMissing = "KubeAPIServerEndpointsMissing"
)

// reconcileKubeAPIServerEndpoints deploys the GlobalNetworkSet containing the IP addresses of the shoot's
// kube-apiserver endpoint, or removes it if the feature is disabled. A kube-apiserver exposed via a hostname is
// reported as a configuration problem, since only disabling the feature can resolve it. Addresses which are merely not
// published yet are no error at all: failing there would block the shoot's reconciliation flow, including the initial
// creation.
//
// The GlobalNetworkSet CRD is part of the calico chart, i.e. of another managed resource which the
// gardener-resource-manager applies independently of this one. This one may therefore report ResourcesApplied=False on
// a newly created shoot until the CRD is established, which is tolerated, see docs/operations/operations.md.
func (a *actuator) reconcileKubeAPIServerEndpoints(
	ctx context.Context,
	log logr.Logger,
	network *extensionsv1alpha1.Network,
	networkConfig *calicov1alpha1.NetworkConfig,
	cluster *extensionscontroller.Cluster,
) error {
	var config *calicov1alpha1.KubeAPIServerEndpoints
	if networkConfig != nil {
		config = networkConfig.KubeAPIServerEndpoints
	}

	if !apiserverendpoints.Enabled(config, a.kubeAPIServerEndpointsConfig) {
		return managedresources.Delete(ctx, a.client, network.Namespace, KubeAPIServerEndpointsManagedResourceName, true)
	}

	endpoints, err := apiserverendpoints.Collect(ctx, a.apiReader, network.Namespace, cluster)
	switch {
	case err == nil:
	case errors.Is(err, apiserverendpoints.ErrUnsupportedHostname):
		// Nothing but disabling the feature can resolve this, so fail loudly rather than deploying nothing: the shoot
		// owner asked for a GlobalNetworkSet which this landscape cannot provide.
		return v1beta1helper.NewErrorWithCodes(fmt.Errorf("%w, hence no GlobalNetworkSet can be deployed - unset "+
			"`kubeAPIServerEndpoints.enabled` for this shoot or disable it landscape-wide", err),
			gardencorev1beta1.ErrorConfigurationProblem)
	case errors.Is(err, apiserverendpoints.ErrNoIPAddress):
		return a.reportUndeterminableKubeAPIServerEndpoints(ctx, log, network, err)
	default:
		return err
	}

	globalNetworkSet, err := apiserverendpoints.RenderGlobalNetworkSet(endpoints)
	if err != nil {
		return fmt.Errorf("could not render the kube-apiserver GlobalNetworkSet: %w", err)
	}

	log.V(1).Info("Deploying the kube-apiserver GlobalNetworkSet", "nets", endpoints.CIDRs, "source", endpoints.Source)

	return managedresources.CreateForShoot(ctx, a.client, network.Namespace, KubeAPIServerEndpointsManagedResourceName,
		managedResourceOrigin, false, map[string][]byte{kubeAPIServerEndpointsDataKey: globalNetworkSet})
}

// reportUndeterminableKubeAPIServerEndpoints records an event on the Network resource, because the reconciliation
// deliberately succeeds and a log message in the extension pod alone would leave everybody looking at the shoot without
// an indication why policies referring to the GlobalNetworkSet do not work.
func (a *actuator) reportUndeterminableKubeAPIServerEndpoints(
	ctx context.Context,
	log logr.Logger,
	network *extensionsv1alpha1.Network,
	reason error,
) error {
	deployed, err := a.globalNetworkSetDeployed(ctx, network.Namespace)
	if err != nil {
		return err
	}

	name := calico.KubeAPIServerEndpointsName

	if deployed {
		log.Info("Could not determine the kube-apiserver endpoints, keeping the previously deployed GlobalNetworkSet",
			"name", name, "reason", reason.Error())
		a.recorder.Eventf(network, nil, corev1.EventTypeWarning, EventKubeAPIServerEndpointsOutdated,
			gardencorev1beta1.EventActionReconcile,
			"Could not determine the kube-apiserver endpoints (%s). The GlobalNetworkSet %q still holds the addresses "+
				"of the last successful reconciliation and may be outdated.", reason, name)

		return nil
	}

	log.Info("Could not determine the kube-apiserver endpoints, no GlobalNetworkSet is deployed",
		"name", name, "reason", reason.Error())
	a.recorder.Eventf(network, nil, corev1.EventTypeWarning, EventKubeAPIServerEndpointsMissing,
		gardencorev1beta1.EventActionReconcile,
		"Could not determine the kube-apiserver endpoints (%s). The GlobalNetworkSet %q is not deployed, hence Calico "+
			"policies referring to it match no address and block traffic to the kube-apiserver.", reason, name)

	return nil
}

func (a *actuator) globalNetworkSetDeployed(ctx context.Context, namespace string) (bool, error) {
	managedResource := &resourcesv1alpha1.ManagedResource{}

	if err := a.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: KubeAPIServerEndpointsManagedResourceName}, managedResource); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("could not check whether the kube-apiserver GlobalNetworkSet is deployed: %w", err)
	}

	return true, nil
}
