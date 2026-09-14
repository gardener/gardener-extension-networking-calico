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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	typhaRestartedAtAnnotation = "networking.calico.extensions.gardener.cloud/typha-migration-restart-at"
	controlPlaneHAAnnotation   = "networking.calico.extensions.gardener.cloud/control-plane-ha"
)

// Restore implements Network.Actuator.
// Before reconciling, it waits for the shoot API server watch cache to warm up and then annotates
// the Network resource with a restart timestamp. The chart rendering picks up this annotation and
// includes it in the calico-typha pod template, causing a rolling restart via the ManagedResource.
// Storing the timestamp on the Network resource means subsequent reconciles keep the same annotation
// value and do not trigger further restarts.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, network *extensionsv1alpha1.Network, cluster *extensionscontroller.Cluster) error {
	typhaEnabled := isTyphaEnabled(network)

	if typhaEnabled {
		if _, alreadySet := network.Annotations[typhaRestartedAtAnnotation]; !alreadySet {
			shootClient, err := a.getShootClient(ctx, cluster)
			if err != nil {
				return fmt.Errorf("failed to get shoot client for calico-typha restart: %w", err)
			}

			log.Info("Checking shoot API server watch cache before restarting calico-typha")
			if err := ensureAPIServerWatchCacheWarm(ctx, shootClient); err != nil {
				return fmt.Errorf("shoot API server watch cache not yet warm, retrying: %w", err)
			}

			patch := client.MergeFrom(network.DeepCopy())
			if network.Annotations == nil {
				network.Annotations = map[string]string{}
			}
			network.Annotations[typhaRestartedAtAnnotation] = time.Now().UTC().Format(time.RFC3339)
			if err := a.client.Patch(ctx, network, patch); err != nil {
				return fmt.Errorf("failed to annotate Network resource for calico-typha restart: %w", err)
			}
			log.Info("Annotated Network resource to trigger calico-typha rolling restart after control plane restore")
		}
	}

	return a.Reconcile(ctx, log, network, cluster)
}

// isTyphaEnabled returns true unless the NetworkConfig explicitly disables Typha.
func isTyphaEnabled(network *extensionsv1alpha1.Network) bool {
	if network.Spec.ProviderConfig == nil || network.Spec.ProviderConfig.Raw == nil {
		return true
	}
	networkConfig, err := CalicoNetworkConfigFromNetworkResource(network)
	if err != nil {
		return true
	}
	return networkConfig.Typha == nil || networkConfig.Typha.Enabled
}

// ensureAPIServerWatchCacheWarm checks once whether the shoot API server's watch cache for
// WorkloadEndpoints (a Calico CRD) is warm. CRD caches are populated asynchronously after
// the API server passes its readiness probe, so a non-zero resourceVersion confirms Typha
// can safely reconnect. Returns an error if not yet ready so the reconciler requeues.
func ensureAPIServerWatchCacheWarm(ctx context.Context, shootClient client.Client) error {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "crd.projectcalico.org",
		Version: "v1",
		Kind:    "WorkloadEndpointList",
	})
	if err := shootClient.List(ctx, list); err != nil {
		return fmt.Errorf("WorkloadEndpoint list not yet available: %w", err)
	}
	if rv := list.GetResourceVersion(); rv == "" || rv == "0" {
		return fmt.Errorf("API server watch cache not yet warm for WorkloadEndpoints (resourceVersion=%q)", rv)
	}
	return nil
}
