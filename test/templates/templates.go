// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package templates

const (

	// NetworkTestName is the name of a network ping test for calico
	NetworkTestName      = "network-test.yaml.tpl"
	NetworkTestNamespace = "default"

	// NetworkProbeMeshName is the name of the per-node HTTP server + client probe
	// mesh template used by the SeamlessOverlaySwitch integration test.
	NetworkProbeMeshName = "network-probe-mesh.yaml.tpl"
)
