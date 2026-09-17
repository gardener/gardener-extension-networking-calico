// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1/helper"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

// Restore implements Network.Actuator.
// Before reconciling, it waits for the shoot API server watch cache to warm up and then annotates
// the Network resource with a fresh timestamp. The chart rendering picks up this annotation and
// includes it in the pod template of either calico-typha (when Typha is enabled) or the
// calico-node DaemonSet (when Typha is disabled), causing a rolling restart via the ManagedResource.
// The timestamp changes on every CPM, so the pod template annotation diff triggers the restart
// naturally. The annotation is kept across reconciles so subsequent Reconcile calls render the same
// value and do not trigger further restarts.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	typhaEnabled, err := isTyphaEnabled(network)
	if err != nil {
		return fmt.Errorf("failed to determine Typha state for CPM restart: %w", err)
	}

	component := "calico-node"
	restartedAtKey := calico.AnnotationCalicoNodeRestartedAt
	restartReasonKey := calico.AnnotationCalicoNodeRestartReason
	if typhaEnabled {
		component = "calico-typha"
		restartedAtKey = calico.AnnotationTyphaRestartedAt
		restartReasonKey = calico.AnnotationTyphaRestartReason
	}

	shootClient, err := a.getShootClient(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to get shoot client for %s restart: %w", component, err)
	}

	log.Info("Waiting for shoot API server watch cache to be ready before restarting " + component)
	if err := ensureAPIServerWatchCacheWarm(ctx, shootClient); err != nil {
		// Expected during shoot API server startup after a control plane migration;
		// the reconciler will requeue until the cache is warm.
		return fmt.Errorf("shoot API server not yet ready after control plane migration, requeueing: %w", err)
	}

	patch := client.MergeFrom(network.DeepCopy())
	if network.Annotations == nil {
		network.Annotations = map[string]string{}
	}
	network.Annotations[restartedAtKey] = time.Now().UTC().Format(time.RFC3339)
	network.Annotations[restartReasonKey] = calico.RestartReasonCPM
	if err := a.client.Patch(ctx, network, patch); err != nil {
		return fmt.Errorf("failed to annotate Network resource for %s restart: %w", component, err)
	}
	log.Info("Annotated Network resource to trigger " + component + " rolling restart after control plane restore")

	return a.Reconcile(ctx, log, network, cluster)
}

// isTyphaEnabled returns true unless the NetworkConfig explicitly disables Typha.
// Returns an error if the ProviderConfig cannot be decoded, so callers can fail
// fast rather than silently routing to the wrong restart path.
func isTyphaEnabled(network *extensionsv1alpha1.Network) (bool, error) {
	if network.Spec.ProviderConfig == nil || network.Spec.ProviderConfig.Raw == nil {
		return true, nil
	}
	networkConfig, err := helper.CalicoNetworkConfigFromNetworkResource(network)
	if err != nil {
		return false, fmt.Errorf("failed to decode NetworkConfig: %w", err)
	}
	return networkConfig.Typha == nil || networkConfig.Typha.Enabled, nil
}

// ensureAPIServerWatchCacheWarm checks once whether the shoot API server's watch cache for
// FelixConfigurations is warm. The API server warms CRD watch caches asynchronously after
// passing its readiness probe, so a non-zero resourceVersion on a Calico CRD confirms that
// the caches Typha depends on are ready. FelixConfiguration is deployed by this extension's
// chart and survives CPM via KeepObjects, so it is always available as a probe target.
// Returns an error if not yet ready so the reconciler requeues.
func ensureAPIServerWatchCacheWarm(ctx context.Context, shootClient client.Client) error {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "crd.projectcalico.org",
		Version: "v1",
		Kind:    "FelixConfigurationList",
	})
	if err := shootClient.List(ctx, list); err != nil {
		if apimeta.IsNoMatchError(err) {
			// CRD not yet installed on the shoot; the warm-cache concern does not apply.
			// Proceeding lets Restore call Reconcile, which will install it.
			return nil
		}
		return fmt.Errorf("FelixConfiguration list not yet available: %w", err)
	}
	if rv := list.GetResourceVersion(); rv == "" || rv == "0" {
		return fmt.Errorf("API server watch cache not yet warm for Calico CRDs (resourceVersion=%q)", rv)
	}
	return nil
}
