// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTableSubtree("Network Extension Tests", Label("Network"), func(shootFunc func() *gardencorev1beta1.Shoot) {
	f := defaultShootCreationFramework()
	f.Shoot = shootFunc()

	It("Create Shoot, Test Network, Force Delete Shoot", Label("force-delete"), func() {
		By("Create Shoot")
		ctx, cancel := context.WithTimeout(parentCtx, 15*time.Minute)
		defer cancel()
		Expect(f.CreateShootAndWaitForCreation(ctx, false)).To(Succeed())
		f.Verify()

		ctx, cancel = context.WithTimeout(parentCtx, 15*time.Minute)
		defer cancel()
		succeeded := testNetwork(ctx, f)

		By("Wait for Shoot to be force-deleted")
		ctx, cancel = context.WithTimeout(parentCtx, 10*time.Minute)
		defer cancel()
		Expect(f.ForceDeleteShootAndWaitForDeletion(ctx, f.Shoot)).To(Succeed())

		By("Network Test status")
		Expect(succeeded).To(BeTrue())
	})
},
	Entry("default", func() *gardencorev1beta1.Shoot { return defaultShoot("e2e-force-del") }),
	Entry("ebpf", func() *gardencorev1beta1.Shoot { return ebpfShoot("e2e-fd-bpf") }),
)
