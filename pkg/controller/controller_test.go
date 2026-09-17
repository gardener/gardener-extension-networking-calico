// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"encoding/json"
	"errors"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

var _ = Describe("isTyphaEnabled", func() {
	It("returns true when ProviderConfig is nil", func() {
		network := &extensionsv1alpha1.Network{}
		enabled, err := isTyphaEnabled(network)
		Expect(err).NotTo(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("returns true when ProviderConfig.Raw is nil", func() {
		network := &extensionsv1alpha1.Network{
			Spec: extensionsv1alpha1.NetworkSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					ProviderConfig: &runtime.RawExtension{},
				},
			},
		}
		enabled, err := isTyphaEnabled(network)
		Expect(err).NotTo(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("returns true when Typha field is absent in config", func() {
		enabled, err := isTyphaEnabled(networkWithTyphaConfig(nil))
		Expect(err).NotTo(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("returns true when Typha.Enabled is true", func() {
		enabled, err := isTyphaEnabled(networkWithTyphaConfig(&calicov1alpha1.Typha{Enabled: true}))
		Expect(err).NotTo(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("returns false when Typha.Enabled is false", func() {
		enabled, err := isTyphaEnabled(networkWithTyphaConfig(&calicov1alpha1.Typha{Enabled: false}))
		Expect(err).NotTo(HaveOccurred())
		Expect(enabled).To(BeFalse())
	})

	It("returns an error when ProviderConfig contains invalid JSON", func() {
		network := &extensionsv1alpha1.Network{
			Spec: extensionsv1alpha1.NetworkSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					ProviderConfig: &runtime.RawExtension{Raw: []byte("not-json")},
				},
			},
		}
		_, err := isTyphaEnabled(network)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("ensureAPIServerWatchCacheWarm", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("returns an error when the List call fails", func() {
		shootClient := interceptor.NewClient(fakeclient.NewClientBuilder().Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("connection refused")
			},
		})
		Expect(ensureAPIServerWatchCacheWarm(ctx, shootClient)).To(MatchError(ContainSubstring("connection refused")))
	})

	It("returns an error when resourceVersion is empty", func() {
		shootClient := interceptor.NewClient(fakeclient.NewClientBuilder().Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
				list.(*unstructured.UnstructuredList).SetResourceVersion("")
				return nil
			},
		})
		Expect(ensureAPIServerWatchCacheWarm(ctx, shootClient)).To(MatchError(ContainSubstring("not yet warm")))
	})

	It("returns an error when resourceVersion is zero", func() {
		shootClient := interceptor.NewClient(fakeclient.NewClientBuilder().Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
				list.(*unstructured.UnstructuredList).SetResourceVersion("0")
				return nil
			},
		})
		Expect(ensureAPIServerWatchCacheWarm(ctx, shootClient)).To(MatchError(ContainSubstring("not yet warm")))
	})

	It("returns nil when resourceVersion is non-zero", func() {
		shootClient := interceptor.NewClient(fakeclient.NewClientBuilder().Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
				list.(*unstructured.UnstructuredList).SetResourceVersion("12345")
				return nil
			},
		})
		Expect(ensureAPIServerWatchCacheWarm(ctx, shootClient)).To(Succeed())
	})
})

var _ = Describe("handleHATransition", func() {
	var (
		ctx        context.Context
		log        logr.Logger
		scheme     *runtime.Scheme
		nonHAShoot *gardencorev1beta1.Shoot
		haShoot    *gardencorev1beta1.Shoot
	)

	BeforeEach(func() {
		ctx = context.Background()
		log = logr.Discard()

		scheme = runtime.NewScheme()
		utilruntime.Must(extensionsv1alpha1.AddToScheme(scheme))

		nonHAShoot = &gardencorev1beta1.Shoot{}
		haShoot = &gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				ControlPlane: &gardencorev1beta1.ControlPlane{
					HighAvailability: &gardencorev1beta1.HighAvailability{
						FailureTolerance: gardencorev1beta1.FailureTolerance{
							Type: gardencorev1beta1.FailureToleranceTypeNode,
						},
					},
				},
			},
		}
	})

	newActuator := func(network *extensionsv1alpha1.Network) *actuator {
		return &actuator{
			client: fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build(),
		}
	}

	newNetwork := func(annotations map[string]string) *extensionsv1alpha1.Network {
		return &extensionsv1alpha1.Network{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "network",
				Namespace:   "shoot--project--name",
				Annotations: annotations,
			},
		}
	}

	clusterFor := func(shoot *gardencorev1beta1.Shoot) *extensionscontroller.Cluster {
		return &extensionscontroller.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "shoot--project--name"},
			Shoot:      shoot,
		}
	}

	It("is idempotent when control-plane-ha annotation already matches non-HA state", func() {
		network := newNetwork(map[string]string{calico.AnnotationControlPlaneHA: "false"})
		act := newActuator(network)
		Expect(act.handleHATransition(ctx, log, network, clusterFor(nonHAShoot))).To(Succeed())
		Expect(network.Annotations).NotTo(HaveKey(calico.AnnotationTyphaRestartedAt))
	})

	It("is idempotent when control-plane-ha annotation already matches HA state", func() {
		network := newNetwork(map[string]string{calico.AnnotationControlPlaneHA: "true"})
		act := newActuator(network)
		Expect(act.handleHATransition(ctx, log, network, clusterFor(haShoot))).To(Succeed())
		Expect(network.Annotations).NotTo(HaveKey(calico.AnnotationTyphaRestartedAt))
	})

	It("sets annotation to false on first reconcile of a non-HA cluster", func() {
		network := newNetwork(nil)
		act := newActuator(network)
		Expect(act.handleHATransition(ctx, log, network, clusterFor(nonHAShoot))).To(Succeed())
		Expect(network.Annotations).To(HaveKeyWithValue(calico.AnnotationControlPlaneHA, "false"))
		Expect(network.Annotations).NotTo(HaveKey(calico.AnnotationTyphaRestartedAt))
	})

	It("sets annotation to true on first reconcile of an HA cluster without triggering a restart", func() {
		network := newNetwork(nil)
		act := newActuator(network)
		Expect(act.handleHATransition(ctx, log, network, clusterFor(haShoot))).To(Succeed())
		Expect(network.Annotations).To(HaveKeyWithValue(calico.AnnotationControlPlaneHA, "true"))
		Expect(network.Annotations).NotTo(HaveKey(calico.AnnotationTyphaRestartedAt))
	})

	It("updates HA annotation without restart when Typha is disabled during non-HA->HA transition", func() {
		network := newNetwork(map[string]string{calico.AnnotationControlPlaneHA: "false"})
		network.Spec.ProviderConfig = networkWithTyphaConfig(&calicov1alpha1.Typha{Enabled: false}).Spec.ProviderConfig
		act := newActuator(network)
		Expect(act.handleHATransition(ctx, log, network, clusterFor(haShoot))).To(Succeed())
		Expect(network.Annotations).To(HaveKeyWithValue(calico.AnnotationControlPlaneHA, "true"))
		Expect(network.Annotations).NotTo(HaveKey(calico.AnnotationTyphaRestartedAt))
	})

	It("returns an error when Typha restart is needed but the shoot client cannot be obtained", func() {
		network := newNetwork(map[string]string{calico.AnnotationControlPlaneHA: "false"})
		act := newActuator(network)
		err := act.handleHATransition(ctx, log, network, clusterFor(haShoot))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get shoot client"))
	})
})

// networkWithTyphaConfig builds a Network whose ProviderConfig encodes the given Typha settings.
// Passing nil omits the typha field entirely.
func networkWithTyphaConfig(typha *calicov1alpha1.Typha) *extensionsv1alpha1.Network {
	cfg := map[string]interface{}{
		"apiVersion": calicov1alpha1.SchemeGroupVersion.String(),
		"kind":       "NetworkConfig",
	}
	if typha != nil {
		cfg["typha"] = map[string]interface{}{"enabled": typha.Enabled}
	}
	raw, _ := json.Marshal(cfg)
	return &extensionsv1alpha1.Network{
		Spec: extensionsv1alpha1.NetworkSpec{
			DefaultSpec: extensionsv1alpha1.DefaultSpec{
				ProviderConfig: &runtime.RawExtension{Raw: raw},
			},
		},
	}
}

var _ = Describe("Restore (calico-node path)", func() {
	var (
		ctx    context.Context
		log    logr.Logger
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		log = logr.Discard()
		scheme = runtime.NewScheme()
		utilruntime.Must(extensionsv1alpha1.AddToScheme(scheme))
	})

	newNetworkNoTypha := func(annotations map[string]string) *extensionsv1alpha1.Network {
		n := networkWithTyphaConfig(&calicov1alpha1.Typha{Enabled: false})
		n.Name = "network"
		n.Namespace = "shoot--project--name"
		n.Annotations = annotations
		return n
	}

	newActuator := func(network *extensionsv1alpha1.Network) *actuator {
		return &actuator{
			client: fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(network).Build(),
		}
	}

	clusterFor := func() *extensionscontroller.Cluster {
		return &extensionscontroller.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "shoot--project--name"},
			Shoot:      &gardencorev1beta1.Shoot{},
		}
	}

	It("returns an error when the shoot client cannot be obtained (cache check blocked)", func() {
		network := newNetworkNoTypha(nil)
		act := newActuator(network)
		err := act.Restore(ctx, log, network, clusterFor())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get shoot client for calico-node restart"))
	})

	It("returns an error when annotation is already set from a prior CPM (fresh timestamp always attempted)", func() {
		// With no "already set" guard, every Restore call attempts the restart.
		// A pre-existing annotation does not suppress the restart — the shoot client
		// error confirms the restart path is always entered.
		network := newNetworkNoTypha(map[string]string{
			calico.AnnotationCalicoNodeRestartedAt: "2026-09-01T10:00:00Z",
		})
		act := newActuator(network)
		err := act.Restore(ctx, log, network, clusterFor())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get shoot client for calico-node restart"))
	})
})
