// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
)

var _ = Describe("managed resource lifecycle", func() {
	const namespace = "shoot--foo--bar"

	var (
		ctx     = context.Background()
		network = &extensionsv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{Name: "calico-network", Namespace: namespace}}
		cluster = &extensionscontroller.Cluster{Shoot: &gardencorev1beta1.Shoot{}}

		newClientWithManagedResource = func() client.WithWatch {
			return fake.NewClientBuilder().WithScheme(testScheme).WithObjects(&resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: CalicoConfigManagedResourceName, Namespace: namespace},
			}).Build()
		}

		get = func(c client.Client, name string) (*resourcesv1alpha1.ManagedResource, error) {
			mr := &resourcesv1alpha1.ManagedResource{}
			err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, mr)
			return mr, err
		}
	)

	It("should delete the managed resource on Delete", func() {
		c := newClientWithManagedResource()

		Expect((&actuator{client: c, apiReader: c}).Delete(ctx, logr.Discard(), network, cluster)).To(Succeed())

		_, err := get(c, CalicoConfigManagedResourceName)
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "the managed resource should be gone")
	})

	It("should delete the managed resource on ForceDelete", func() {
		c := newClientWithManagedResource()

		Expect((&actuator{client: c, apiReader: c}).ForceDelete(ctx, logr.Discard(), network, cluster)).To(Succeed())

		_, err := get(c, CalicoConfigManagedResourceName)
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "the managed resource should be gone")
	})

	It("should succeed on Delete if the managed resource does not exist", func() {
		c := fake.NewClientBuilder().WithScheme(testScheme).Build()

		Expect((&actuator{client: c, apiReader: c}).Delete(ctx, logr.Discard(), network, cluster)).To(Succeed())
	})

	It("should succeed on Migrate if the managed resource does not exist", func() {
		c := fake.NewClientBuilder().WithScheme(testScheme).Build()

		Expect((&actuator{client: c, apiReader: c}).Migrate(ctx, logr.Discard(), network, cluster)).To(Succeed())
	})

	It("should keep the objects of the managed resource on Migrate", func() {
		c := newClientWithManagedResource()
		keepObjects := map[string]*bool{}

		// Record the keepObjects flag before Migrate deletes the managed resources.
		c = interceptor.NewClient(c, interceptor.Funcs{
			Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if mr, ok := obj.(*resourcesv1alpha1.ManagedResource); ok {
					current := &resourcesv1alpha1.ManagedResource{}
					if err := cl.Get(ctx, client.ObjectKeyFromObject(mr), current); err == nil {
						keepObjects[mr.Name] = current.Spec.KeepObjects
					}
				}
				return cl.Delete(ctx, obj, opts...)
			},
		})

		Expect((&actuator{client: c, apiReader: c}).Migrate(ctx, logr.Discard(), network, cluster)).To(Succeed())

		Expect(keepObjects).To(HaveKeyWithValue(CalicoConfigManagedResourceName, ptr.To(true)))
	})
})

var _ = Describe("#desiredKubeAPIServerCIDRs", func() {
	const namespace = "shoot--foo--bar"

	var (
		ctx = context.Background()

		enabled = &calicov1alpha1.NetworkConfig{
			KubeAPIServerGlobalNetworkSet: &calicov1alpha1.KubeAPIServerGlobalNetworkSet{Enabled: ptr.To(true)},
		}

		dnsRecord = &extensionsv1alpha1.DNSRecord{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-internal",
				Namespace: namespace,
				Labels: map[string]string{
					v1beta1constants.LabelRole:  v1beta1constants.LabelDNSRecordInternal,
					v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane,
				},
			},
			Spec: extensionsv1alpha1.DNSRecordSpec{
				Name:       "api.foo.bar.example.com",
				RecordType: extensionsv1alpha1.DNSRecordTypeA,
				Values:     []string{"34.107.12.34"},
			},
		}

		newCluster = func(hibernationEnabled, isHibernated bool) *extensionscontroller.Cluster {
			return &extensionscontroller.Cluster{Shoot: &gardencorev1beta1.Shoot{
				Spec:   gardencorev1beta1.ShootSpec{Hibernation: &gardencorev1beta1.Hibernation{Enabled: ptr.To(hibernationEnabled)}},
				Status: gardencorev1beta1.ShootStatus{IsHibernated: isHibernated},
			}}
		}

		newActuator = func(c client.Reader) *actuator {
			return &actuator{apiReader: c}
		}

		clientWithDNSRecord = func() client.WithWatch {
			return fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord).Build()
		}
	)

	It("should return nothing if the feature is disabled", func() {
		// The client rejects every read, so a lookup would fail the test rather than return nothing.
		c := interceptor.NewClient(clientWithDNSRecord(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("the DNSRecords must not be read")
			},
		})

		cidrs, err := newActuator(c).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(false, false), nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(BeEmpty())
	})

	It("should return the addresses if the feature is enabled and the shoot is awake", func() {
		cidrs, err := newActuator(clientWithDNSRecord()).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(false, false), enabled)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(ConsistOf("34.107.12.34/32"))
	})

	It("should fail if the addresses cannot be determined and the shoot is awake", func() {
		c := fake.NewClientBuilder().WithScheme(testScheme).Build()

		_, err := newActuator(c).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(false, false), enabled)

		Expect(err).To(MatchError(ContainSubstring("do not publish an address yet")))
	})

	It("should return nothing without reading the DNSRecords if the shoot is hibernated", func() {
		// gardenlet destroyed the DNSRecords, so reading them must not even be attempted.
		c := interceptor.NewClient(clientWithDNSRecord(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("the DNSRecords must not be read")
			},
		})

		cidrs, err := newActuator(c).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(true, true), enabled)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(BeEmpty())
	})

	It("should still publish the addresses while the shoot is going into hibernation", func() {
		// The DNSRecords are destroyed only after the control plane was hibernated, so they are still there.
		cidrs, err := newActuator(clientWithDNSRecord()).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(true, false), enabled)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(ConsistOf("34.107.12.34/32"))
	})

	It("should publish the addresses again while the shoot is waking up", func() {
		cidrs, err := newActuator(clientWithDNSRecord()).desiredKubeAPIServerCIDRs(ctx, namespace, newCluster(false, true), enabled)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(ConsistOf("34.107.12.34/32"))
	})

	It("should treat a shoot without a hibernation section as awake", func() {
		cluster := &extensionscontroller.Cluster{Shoot: &gardencorev1beta1.Shoot{}}

		cidrs, err := newActuator(clientWithDNSRecord()).desiredKubeAPIServerCIDRs(ctx, namespace, cluster, enabled)

		Expect(err).NotTo(HaveOccurred())
		Expect(cidrs).To(ConsistOf("34.107.12.34/32"))
	})
})
