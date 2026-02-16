// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

const (
	// NonPrivilegedCalicoNode runs the long-lived calico-node container in non-privileged and non-root mode.
	// owner @ialidzhikov
	// alpha: v1.27.0
	NonPrivilegedCalicoNode featuregate.Feature = "NonPrivilegedCalicoNode"
	// SeamlessOverlaySwitch validates node routes before disabling overlay networking.
	// owner @docktofuture
	// alpha: v1.38.0
	SeamlessOverlaySwitch featuregate.Feature = "SeamlessOverlaySwitch"
)

var (
	// FeatureGate is a shared global FeatureGate for networking-calico extension flags.
	FeatureGate  = featuregate.NewFeatureGate()
	featureGates = map[featuregate.Feature]featuregate.FeatureSpec{
		NonPrivilegedCalicoNode: {Default: false, PreRelease: featuregate.Alpha},
		SeamlessOverlaySwitch:   {Default: false, PreRelease: featuregate.Alpha},
	}
)

// RegisterFeatureGates registers the feature gates of the networking-calico extension.
func RegisterFeatureGates() {
	runtime.Must(FeatureGate.Add(featureGates))
}
