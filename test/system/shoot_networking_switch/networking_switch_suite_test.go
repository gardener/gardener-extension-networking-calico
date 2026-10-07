// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shootnetworkingswitch_test

import (
	"testing"

	"github.com/gardener/gardener/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func init() {
	framework.RegisterShootFrameworkFlags()
}

func TestNetworkingSwitch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Calico Overlay-Native Seamless Switch Test Suite")
}
