// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
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
