// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator_test

import (
	"github.com/gardener/gardener/pkg/apis/core"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	. "github.com/gardener/gardener-extension-networking-calico/pkg/admission/validator"
)

var _ = Describe("Shoot validator", func() {
	Describe("#CalicoPredicate", func() {
		var (
			p     predicate.Predicate
			shoot *core.Shoot
		)

		BeforeEach(func() {
			p = CalicoPredicate()
			shoot = &core.Shoot{}
		})

		It("should return true because the networking type is calico", func() {
			shoot.Spec.Networking = &core.Networking{
				Type: pointer.String("calico"),
			}

			Expect(p.Create(event.CreateEvent{Object: shoot})).To(BeTrue())
			Expect(p.Update(event.UpdateEvent{ObjectNew: shoot})).To(BeTrue())
			Expect(p.Delete(event.DeleteEvent{Object: shoot})).To(BeTrue())
			Expect(p.Generic(event.GenericEvent{Object: shoot})).To(BeTrue())
		})

		It("should return false because the networking type is not calico", func() {
			shoot.Spec.Networking = &core.Networking{
				Type: pointer.String("other-provider"),
			}

			Expect(p.Create(event.CreateEvent{Object: shoot})).To(BeFalse())
			Expect(p.Update(event.UpdateEvent{ObjectNew: shoot})).To(BeFalse())
			Expect(p.Delete(event.DeleteEvent{Object: shoot})).To(BeFalse())
			Expect(p.Generic(event.GenericEvent{Object: shoot})).To(BeFalse())
		})

		It("should return false because the networking is nil", func() {
			Expect(p.Create(event.CreateEvent{Object: shoot})).To(BeFalse())
			Expect(p.Update(event.UpdateEvent{ObjectNew: shoot})).To(BeFalse())
			Expect(p.Delete(event.DeleteEvent{Object: shoot})).To(BeFalse())
			Expect(p.Generic(event.GenericEvent{Object: shoot})).To(BeFalse())
		})

		It("should return false because the networking type is nil", func() {
			shoot.Spec.Networking = &core.Networking{}

			Expect(p.Create(event.CreateEvent{Object: shoot})).To(BeFalse())
			Expect(p.Update(event.UpdateEvent{ObjectNew: shoot})).To(BeFalse())
			Expect(p.Delete(event.DeleteEvent{Object: shoot})).To(BeFalse())
			Expect(p.Generic(event.GenericEvent{Object: shoot})).To(BeFalse())
		})
	})
})
