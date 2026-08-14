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
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	apisconfig "github.com/gardener/gardener-extension-networking-calico/pkg/apis/config"
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"
)

var _ = Describe("#reconcileKubeAPIServerEndpoints", func() {
	const namespace = "shoot--foo--bar"

	var (
		ctx      = context.Background()
		network  = &extensionsv1alpha1.Network{ObjectMeta: metav1.ObjectMeta{Name: "calico-network", Namespace: namespace}}
		recorder *fakeRecorder

		newActuator = func(c client.Client, operatorConfig *bool) *actuator {
			recorder = &fakeRecorder{}
			a := &actuator{client: c, apiReader: c, recorder: recorder}
			if operatorConfig != nil {
				a.kubeAPIServerEndpointsConfig = &apisconfig.KubeAPIServerEndpointsConfiguration{Enabled: operatorConfig}
			}
			return a
		}

		networkConfig = func(endpoints *calicov1alpha1.KubeAPIServerEndpoints) *calicov1alpha1.NetworkConfig {
			return &calicov1alpha1.NetworkConfig{KubeAPIServerEndpoints: endpoints}
		}

		enabled = networkConfig(&calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(true)})

		dnsRecord = func(recordType extensionsv1alpha1.DNSRecordType, values ...string) *extensionsv1alpha1.DNSRecord {
			return &extensionsv1alpha1.DNSRecord{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-internal",
					Namespace: namespace,
					Labels: map[string]string{
						v1beta1constants.LabelRole:  v1beta1constants.LabelDNSRecordInternal,
						v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane,
					},
				},
				Spec: extensionsv1alpha1.DNSRecordSpec{RecordType: recordType, Values: values},
			}
		}

		cluster = func(addresses ...gardencorev1beta1.ShootAdvertisedAddress) *extensionscontroller.Cluster {
			shoot := &gardencorev1beta1.Shoot{}
			shoot.Status.AdvertisedAddresses = addresses
			return &extensionscontroller.Cluster{Shoot: shoot}
		}

		managedResource = func(c client.Client) (*resourcesv1alpha1.ManagedResource, error) {
			mr := &resourcesv1alpha1.ManagedResource{}
			err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: KubeAPIServerEndpointsManagedResourceName}, mr)
			return mr, err
		}

		globalNetworkSet = func(c client.Client, mr *resourcesv1alpha1.ManagedResource) string {
			ExpectWithOffset(1, mr.Spec.SecretRefs).To(HaveLen(1))
			secret := &corev1.Secret{}
			ExpectWithOffset(1, c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: mr.Spec.SecretRefs[0].Name}, secret)).To(Succeed())
			return string(secret.Data[kubeAPIServerEndpointsDataKey])
		}

		secretNames = func(c client.Client) []string {
			secretList := &corev1.SecretList{}
			ExpectWithOffset(1, c.List(ctx, secretList, client.InNamespace(namespace))).To(Succeed())

			var names []string
			for _, secret := range secretList.Items {
				names = append(names, secret.Name)
			}
			return names
		}

		expectNoEvents = func() {
			ExpectWithOffset(1, recorder.events).To(BeEmpty())
		}
	)

	Context("disabled", func() {
		It("should not deploy anything if neither layer enables it", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")).Build()

			Expect(newActuator(c, nil).reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, networkConfig(nil), cluster())).To(Succeed())

			_, err := managedResource(c)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			expectNoEvents()
		})

		It("should remove an existing managed resource if the providerConfig opts out of the operator default", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(
				dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				&resourcesv1alpha1.ManagedResource{ObjectMeta: metav1.ObjectMeta{Name: KubeAPIServerEndpointsManagedResourceName, Namespace: namespace}},
			).Build()
			a := newActuator(c, ptr.To(true))

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(),
				network, networkConfig(&calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(false)}), cluster())).To(Succeed())

			_, err := managedResource(c)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			expectNoEvents()
		})
	})

	Context("enabled", func() {
		It("should deploy the GlobalNetworkSet from the DNSRecord", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")).Build()
			a := newActuator(c, nil)

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

			mr, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(mr.Labels).To(HaveKeyWithValue("origin", managedResourceOrigin))
			Expect(mr.Spec.KeepObjects).To(Equal(ptr.To(false)))

			Expect(globalNetworkSet(c, mr)).To(SatisfyAll(
				ContainSubstring("apiVersion: crd.projectcalico.org/v1"),
				ContainSubstring("kind: GlobalNetworkSet"),
				ContainSubstring("name: "+calico.KubeAPIServerEndpointsName),
				ContainSubstring(calico.KubeAPIServerEndpointsLabelKey+": "+calico.KubeAPIServerEndpointsLabelValue),
				ContainSubstring("networking.gardener.cloud/source: DNSRecord"),
				ContainSubstring("- 34.107.12.34/32"),
			))
			expectNoEvents()
		})

		It("should deploy the GlobalNetworkSet if only the operator default enables it", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")).Build()
			a := newActuator(c, ptr.To(true))

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, networkConfig(nil), cluster())).To(Succeed())

			mr, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(globalNetworkSet(c, mr)).To(ContainSubstring("- 34.107.12.34/32"))
		})

		It("should deploy dual-stack addresses", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).WithObjects(
				dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34"),
				&extensionsv1alpha1.DNSRecord{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-external",
						Namespace: namespace,
						Labels: map[string]string{
							v1beta1constants.LabelRole:  v1beta1constants.LabelDNSRecordExternal,
							v1beta1constants.GardenRole: v1beta1constants.GardenRoleControlPlane,
						},
					},
					Spec: extensionsv1alpha1.DNSRecordSpec{RecordType: extensionsv1alpha1.DNSRecordTypeAAAA, Values: []string{"2001:db8::1"}},
				},
			).Build()
			a := newActuator(c, nil)

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

			mr, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(globalNetworkSet(c, mr)).To(SatisfyAll(
				ContainSubstring("- 34.107.12.34/32"),
				ContainSubstring("- 2001:db8::1/128"),
			))
		})

		It("should fall back to the advertised addresses if no DNSRecord exists", func() {
			c := fake.NewClientBuilder().WithScheme(testScheme).Build()
			a := newActuator(c, nil)

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled,
				cluster(gardencorev1beta1.ShootAdvertisedAddress{Name: v1beta1constants.AdvertisedAddressUnmanaged, URL: "https://172.18.255.1"}),
			)).To(Succeed())

			mr, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(globalNetworkSet(c, mr)).To(SatisfyAll(
				ContainSubstring("- 172.18.255.1/32"),
				ContainSubstring("networking.gardener.cloud/source: AdvertisedAddresses"),
			))
		})

	})

	Context("repeated reconciliation", func() {
		var (
			c client.Client
			a *actuator
		)

		BeforeEach(func() {
			c = fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")).Build()
			a = newActuator(c, nil)

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())
		})

		It("should not create a new secret if the addresses did not change", func() {
			before, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(secretNames(c)).To(HaveLen(1))

			// The rendered object is hashed into the immutable managed resource secret, so a payload changing per
			// reconciliation would pile up secrets and re-apply the object in every shoot cluster.
			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

			after, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(after.Spec.SecretRefs).To(Equal(before.Spec.SecretRefs))
			Expect(secretNames(c)).To(HaveLen(1))
			Expect(globalNetworkSet(c, after)).To(Equal(globalNetworkSet(c, before)))
		})

		It("should publish the new addresses if they changed", func() {
			before, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())

			Expect(c.Delete(ctx, dnsRecord(extensionsv1alpha1.DNSRecordTypeA))).To(Succeed())
			Expect(c.Create(ctx, dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.99"))).To(Succeed())

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

			after, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(after.Spec.SecretRefs).NotTo(Equal(before.Spec.SecretRefs))
			Expect(globalNetworkSet(c, after)).To(SatisfyAll(
				ContainSubstring("- 34.107.12.99/32"),
				Not(ContainSubstring("- 34.107.12.34/32")),
			))
		})
	})

	Context("the kube-apiserver is exposed via a hostname", func() {
		var (
			c client.WithWatch
			a *actuator
		)

		BeforeEach(func() {
			c = fake.NewClientBuilder().WithScheme(testScheme).WithObjects(
				dnsRecord(extensionsv1alpha1.DNSRecordTypeCNAME, "abc.elb.eu-west-1.amazonaws.com"),
			).Build()
			a = newActuator(c, nil)
		})

		It("should fail with a configuration problem naming the hostname and the way out", func() {
			err := a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())

			Expect(err).To(MatchError(ContainSubstring("the kube-apiserver is exposed via a hostname instead of an IP address")))
			Expect(err).To(MatchError(ContainSubstring("abc.elb.eu-west-1.amazonaws.com")))
			Expect(err).To(MatchError(ContainSubstring("kubeAPIServerEndpoints.enabled")))
			Expect(v1beta1helper.ExtractErrorCodes(err)).To(ConsistOf(gardencorev1beta1.ErrorConfigurationProblem))
		})

		It("should not deploy anything and not record an event, the error is reported via the Network", func() {
			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).NotTo(Succeed())

			_, err := managedResource(c)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			expectNoEvents()
		})

		It("should leave a previously deployed GlobalNetworkSet untouched", func() {
			Expect(c.Create(ctx, &resourcesv1alpha1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{Name: KubeAPIServerEndpointsManagedResourceName, Namespace: namespace},
			})).To(Succeed())

			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).NotTo(Succeed())

			_, err := managedResource(c)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not fail if the feature is disabled", func() {
			Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network,
				networkConfig(&calicov1alpha1.KubeAPIServerEndpoints{Enabled: ptr.To(false)}), cluster())).To(Succeed())
		})
	})

	Context("addresses cannot be determined", func() {
		var (
			c client.WithWatch
			a *actuator
		)

		Context("with a previously deployed GlobalNetworkSet", func() {
			BeforeEach(func() {
				c = fake.NewClientBuilder().WithScheme(testScheme).WithObjects(dnsRecord(extensionsv1alpha1.DNSRecordTypeA, "34.107.12.34")).Build()
				a = newActuator(c, nil)

				// Deploy an initial GlobalNetworkSet which must survive the subsequent reconciliations.
				Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())
			})

			It("should keep the previously deployed GlobalNetworkSet if no address exists at all", func() {
				before, err := managedResource(c)
				Expect(err).NotTo(HaveOccurred())

				Expect(c.Delete(ctx, dnsRecord(extensionsv1alpha1.DNSRecordTypeA))).To(Succeed())

				Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

				after, err := managedResource(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(after.Spec.SecretRefs).To(Equal(before.Spec.SecretRefs))
			})

			It("should record a warning event stating that the set may be outdated", func() {
				Expect(c.Delete(ctx, dnsRecord(extensionsv1alpha1.DNSRecordTypeA))).To(Succeed())

				Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

				Expect(recorder.events).To(ConsistOf(recordedEvent{
					eventType: corev1.EventTypeWarning,
					reason:    EventKubeAPIServerEndpointsOutdated,
					action:    gardencorev1beta1.EventActionReconcile,
					note: "Could not determine the kube-apiserver endpoints (no kube-apiserver IP address could be " +
						"determined: the shoot advertises no kube-apiserver address). The GlobalNetworkSet " +
						`"` + calico.KubeAPIServerEndpointsName + `"` + " still holds the addresses of the last " +
						"successful reconciliation and may be outdated.",
				}))
			})
		})

		Context("without a previously deployed GlobalNetworkSet", func() {
			BeforeEach(func() {
				c = fake.NewClientBuilder().WithScheme(testScheme).Build()
				a = newActuator(c, nil)
			})

			It("should neither fail nor deploy anything", func() {
				Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

				_, err := managedResource(c)
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			})

			It("should record a warning event stating that policies block traffic", func() {
				Expect(a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())).To(Succeed())

				Expect(recorder.events).To(ConsistOf(recordedEvent{
					eventType: corev1.EventTypeWarning,
					reason:    EventKubeAPIServerEndpointsMissing,
					action:    gardencorev1beta1.EventActionReconcile,
					note: "Could not determine the kube-apiserver endpoints (no kube-apiserver IP address could be " +
						"determined: the shoot advertises no kube-apiserver address). The GlobalNetworkSet " +
						`"` + calico.KubeAPIServerEndpointsName + `"` + " is not deployed, hence Calico policies " +
						"referring to it match no address and block traffic to the kube-apiserver.",
				}))
			})

			It("should fail if the existence of the managed resource cannot be determined", func() {
				failing := interceptor.NewClient(c, interceptor.Funcs{
					Get: func(ctx context.Context, cl client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if _, ok := obj.(*resourcesv1alpha1.ManagedResource); ok {
							return errors.New("fake get error")
						}
						return cl.Get(ctx, key, obj, opts...)
					},
				})
				a.client = failing

				err := a.reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())

				Expect(err).To(MatchError(ContainSubstring("could not check whether the kube-apiserver GlobalNetworkSet is deployed")))
			})
		})
	})

	It("should fail if the DNSRecords cannot be read", func() {
		c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(testScheme).Build(), interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return errors.New("fake list error")
			},
		})

		err := newActuator(c, nil).reconcileKubeAPIServerEndpoints(ctx, logr.Discard(), network, enabled, cluster())

		Expect(err).To(MatchError(ContainSubstring("could not read DNSRecords")))
		expectNoEvents()
	})
})

// recordedEvent is an event recorded by fakeRecorder.
type recordedEvent struct {
	eventType, reason, action, note string
}

// fakeRecorder records the events of a single reconciliation. It is used instead of events.FakeRecorder, which discards
// the action - a field the events.k8s.io API requires.
type fakeRecorder struct {
	events []recordedEvent
}

func (f *fakeRecorder) Eventf(_, _ runtime.Object, eventType, reason, action, note string, args ...any) {
	f.events = append(f.events, recordedEvent{
		eventType: eventType,
		reason:    reason,
		action:    action,
		note:      fmt.Sprintf(note, args...),
	})
}
