// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

/*
SeamlessOverlaySwitch system test.

Verifies that the Calico SeamlessOverlaySwitch feature gate prevents cross-node
pod-connectivity loss when a shoot is switched from overlay (IP-in-IP, ipipMode=Always)
to native routing (ipipMode=Never) on an AWS seed.

Prerequisites:
  - A Shoot in overlay mode (overlay.enabled=true, Calico IPPool ipipMode=Always,
    nodes NetworkUnavailable reason=CalicoIsUp).
  - The SeamlessOverlaySwitch feature gate enabled on the networking-calico
    ControllerDeployment.
  - An AWS seed (route controller = aws-custom-route-controller).

See README.md for the full step-by-step walkthrough and the trigger mechanism.
*/

package shootnetworkingswitch_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gcore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gcoreconst "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/gardener/gardener/test/framework"
	"github.com/gardener/gardener/test/utils/access"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8sadmission "k8s.io/api/admissionregistration/v1"
	k8sapps "k8s.io/api/apps/v1"
	k8score "k8s.io/api/core/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gextcalico "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extension-networking-calico/test/templates"
)

// ── timing constants ──────────────────────────────────────────────────────────

const (
	// SwitchTestTimeout is the overall Ginkgo spec timeout.
	SwitchTestTimeout = 30 * time.Minute

	// routeDelay is the sleep duration injected into the route-controller init container.
	// While it sleeps, VPC routes are not created, so the SeamlessOverlaySwitch gate stays
	// engaged. It is a generous ceiling, not a tuned value: the test releases the hazard
	// early (deletes the MAP and the sleeping pod) once the gate-hold assertion is done, so
	// the sleep normally never elapses. It only needs to outlast gateEngagedTimeout (time for
	// Gardener to reconcile far enough to hit the gate) plus gateAssertWindow.
	routeDelay = 18 * time.Minute

	// triggerEngagedTimeout is the maximum time to wait for the route-controller pod with the
	// injected sleep init container to appear on the seed, proving the MAP intercepted the pod
	// CREATE and the hazard is armed.
	triggerEngagedTimeout = 5 * time.Minute

	// triggerPollInterval is the polling cadence while waiting for the trigger to engage.
	triggerPollInterval = 10 * time.Second

	// gateEngagedTimeout is the maximum time to wait for positive evidence that the
	// networking-calico reconcile reached the overlay-switch gate and is holding — surfaced as
	// the gate error on the Network resource's status.lastError. Dominated by how long gardenlet
	// takes to reach the deploy-network step of the shoot reconcile; tune per landscape.
	gateEngagedTimeout = 12 * time.Minute

	// gateEngagedPollInterval is the polling cadence while waiting for the gate to engage.
	gateEngagedPollInterval = 15 * time.Second

	// gateAssertWindow is how long we assert "overlay held" once the gate is confirmed engaged.
	// Must be shorter than routeDelay so the hazard is still active for the entire window.
	gateAssertWindow = 90 * time.Second

	// gateAssertInterval is the polling cadence during the assert window.
	gateAssertInterval = 10 * time.Second
)

// ── name constants ────────────────────────────────────────────────────────────

const (
	// probeNamespace is the namespace for the connectivity probe daemonsets on the shoot.
	probeNamespace = "calico-switch-test"

	// ippoolName is the Calico IPPool CR on every Calico shoot.
	ippoolName = "default-ipv4-ippool"

	// calicoControllerDeploymentName is the name of the networking-calico
	// ControllerDeployment in the garden cluster.
	calicoControllerDeploymentName = "networking-calico"

	// routeControllerDepName is the route-controller Deployment on the seed.
	routeControllerDepName = "aws-custom-route-controller"

	// routeDelayInitContainer is the name of the injected sleep init container.
	routeDelayInitContainer = "route-creation-delay"

	// busyboxImage is the single busybox pin shared by the injected sleep init container
	// (sh -c sleep) and the probe-client (sh -c wget). Pinned (not :latest) for reproducible
	// runs and IfNotPresent pulls; the two uses are distinct but share one version to bump.
	// renovate: datasource=docker
	busyboxImage = "busybox:1.36"

	// mapPolicyBaseName / mapBindingBaseName are the MAP object name prefixes on the seed.
	// routeDelayMAPNames suffixes them with a per-shoot hash so concurrent runs against
	// different shoots on the same seed don't share (and delete) each other's objects.
	mapPolicyBaseName  = "calico-switch-test-route-delay"
	mapBindingBaseName = "calico-switch-test-route-delay-binding"

	// gateHoldErrorSubstring is the substring the networking-calico actuator records on the
	// Network resource's status.lastError while it holds the overlay switch waiting for node
	// routes. Must stay in sync with pkg/controller/actuator_reconcile.go.
	gateHoldErrorSubstring = "waiting for routes to be created"
)

// ── GVK constants ─────────────────────────────────────────────────────────────

var (
	// ippoolGVK is the Calico IPPool CRD served on the shoot cluster.
	ippoolGVK = schema.GroupVersionKind{
		Group:   "crd.projectcalico.org",
		Version: "v1",
		Kind:    "IPPool",
	}

	// controllerDeploymentGVK is the Gardener ControllerDeployment in the garden.
	controllerDeploymentGVK = schema.GroupVersionKind{
		Group:   "core.gardener.cloud",
		Version: "v1",
		Kind:    "ControllerDeployment",
	}

	// networkListGVK is the Gardener Network extension resource list on the seed. The gate's
	// hold state is read from the Network's status.lastError (see gateHoldErrorSubstring).
	networkListGVK = schema.GroupVersionKind{
		Group:   "extensions.gardener.cloud",
		Version: "v1alpha1",
		Kind:    "NetworkList",
	}
)

// ── test entry point ──────────────────────────────────────────────────────────

var _ = Describe("Calico overlay->native seamless switch", func() {
	f := framework.NewShootFramework(nil)

	framework.CIt("should switch overlay->native with zero connectivity loss", func(ctx context.Context) {
		RunTest(ctx, f)
	}, SwitchTestTimeout)
})

// RunTest exercises the SeamlessOverlaySwitch gate end-to-end on an existing shoot.
//
// The test creates a MAP on the seed that delays route creation, triggers the overlay→native
// switch, asserts the gate holds overlay while routes are absent, then lets the MAP sleep
// expire (self-healing) so routes are created and the switch completes gracefully.
func RunTest(ctx context.Context, f *framework.ShootFramework) {
	// Point the framework at this repo's templates directory.
	// Test binary CWD is the package dir; ../.. reaches test/.
	resourceDir, err := filepath.Abs(filepath.Join("..", ".."))
	Expect(err).NotTo(HaveOccurred())
	f.TemplatesDir = filepath.Join(resourceDir, "templates")

	// ── Step 1: gate precondition ────────────────────────────────────────────────
	By("Check gate precondition: SeamlessOverlaySwitch must be enabled on networking-calico ControllerDeployment")
	cd := &unstructured.Unstructured{}
	cd.SetGroupVersionKind(controllerDeploymentGVK)
	Expect(f.GardenClient.Client().Get(ctx, client.ObjectKey{Name: calicoControllerDeploymentName}, cd)).To(
		Succeed(), "ControllerDeployment %q not found in garden", calicoControllerDeploymentName)

	gateEnabled, _, _ := unstructured.NestedBool(cd.Object,
		"helm", "values", "config", "featureGates", "SeamlessOverlaySwitch")
	if !gateEnabled {
		Fail("SeamlessOverlaySwitch feature gate is not enabled on ControllerDeployment " +
			calicoControllerDeploymentName + "; set helm.values.config.featureGates." +
			"SeamlessOverlaySwitch: true on the ControllerDeployment and re-run")
	}
	GinkgoWriter.Println("Gate precondition passed: SeamlessOverlaySwitch is enabled")

	// ── Step 2: create shoot client + start-state assertions ─────────────────────
	By("Create shoot client")
	shootClient, err := access.CreateShootClientFromAdminKubeconfig(ctx, f.GardenClient, f.Shoot)
	Expect(err).NotTo(HaveOccurred())

	By("Assert start-state: IPPool ipipMode=Always (overlay)")
	ipipMode, err := getIPPoolIPIPMode(ctx, shootClient.Client())
	Expect(err).NotTo(HaveOccurred(), "IPPool %q must exist; is this an overlay Calico shoot?", ippoolName)
	Expect(ipipMode).To(Equal("Always"),
		"shoot must be in overlay (ipipMode=Always) before this test; got %q", ipipMode)

	// ── Step 3: deploy probe mesh ─────────────────────────────────────────────────
	By("Deploy probe mesh: server daemonset (empty targets on first pass)")
	Expect(f.RenderAndDeployTemplate(ctx, shootClient, templates.NetworkProbeMeshName,
		struct{ Namespace, Targets, Image string }{probeNamespace, "", busyboxImage}),
	).To(Succeed())

	By("Wait for probe-server daemonset")
	Expect(f.WaitUntilDaemonSetIsRunning(ctx, shootClient.Client(), "probe-server", probeNamespace)).To(Succeed())

	By("Collect server pod IPs")
	serverPods := &k8score.PodList{}
	Expect(shootClient.Client().List(ctx, serverPods,
		client.InNamespace(probeNamespace),
		client.MatchingLabels{"app": "probe-server"},
	)).To(Succeed())
	var serverIPs []string
	serverIPToNode := map[string]string{}
	for _, pod := range serverPods.Items {
		if pod.Status.PodIP != "" {
			serverIPs = append(serverIPs, pod.Status.PodIP)
			serverIPToNode[pod.Status.PodIP] = pod.Spec.NodeName
		}
	}
	Expect(serverIPs).NotTo(BeEmpty(), "no server pod IPs found; server daemonset not ready")
	f.Logger.Info("Collected server pod IPs", "ips", serverIPs)

	By("Redeploy probe mesh with real targets")
	Expect(f.RenderAndDeployTemplate(ctx, shootClient, templates.NetworkProbeMeshName,
		struct{ Namespace, Targets, Image string }{probeNamespace, strings.Join(serverIPs, " "), busyboxImage}),
	).To(Succeed())

	By("Wait for probe-client daemonset")
	Expect(f.WaitUntilDaemonSetIsRunning(ctx, shootClient.Client(), "probe-client", probeNamespace)).To(Succeed())

	// Defer: cleanup probe namespace (runs even on test failure).
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		ns := &k8score.Namespace{ObjectMeta: k8smeta.ObjectMeta{Name: probeNamespace}}
		if delErr := shootClient.Client().Delete(cleanCtx, ns); client.IgnoreNotFound(delErr) != nil {
			f.Logger.Error(delErr, "cleanup: failed to delete probe namespace")
		} else {
			f.Logger.Info("cleanup: deleted probe namespace", "namespace", probeNamespace)
		}
	}()

	// ── Step 4: resolve control-plane namespace ───────────────────────────────────
	// f.Seed / f.SeedClient are already initialized by the ShootFramework's BeforeEach (AddShoot).
	cpNS := f.Shoot.Status.TechnicalID // control-plane namespace from shoot status
	Expect(cpNS).NotTo(BeEmpty(), "shoot status.technicalID (control-plane namespace) must be set")

	// Per-shoot MAP/Binding names so parallel runs against different shoots on the same seed
	// don't delete each other's objects (a re-run of the same shoot reuses the same names).
	mapPolicyName, mapBindingName := routeDelayMAPNames(cpNS)

	// ── Step 5: discover route-controller pod selector ────────────────────────────
	By("Discover route-controller pod selector from Deployment " + routeControllerDepName)
	rcDeploy := &k8sapps.Deployment{}
	Expect(f.SeedClient.Client().Get(ctx,
		client.ObjectKey{Namespace: cpNS, Name: routeControllerDepName}, rcDeploy),
	).To(Succeed(),
		"Deployment %q not found in %q — is this an AWS seed?", routeControllerDepName, cpNS)

	rcLabels := rcDeploy.Spec.Selector.MatchLabels
	Expect(rcLabels).NotTo(BeEmpty(), "route-controller Deployment has empty selector")
	f.Logger.Info("Discovered route-controller pod selector", "labels", rcLabels)

	// ── Step 6: create MAP + Binding on seed ──────────────────────────────────────
	By(fmt.Sprintf("Create MAP %q on seed (injects sleep-%ds init container into route-controller pods)",
		mapPolicyName, int(routeDelay.Seconds())))

	// Clean up any leftover MAP/Binding from a previous failed run before creating new ones.
	deleteRouteDelayMAP(ctx, f, mapPolicyName, mapBindingName)

	mapObj := &k8sadmission.MutatingAdmissionPolicy{
		ObjectMeta: k8smeta.ObjectMeta{
			Name: mapPolicyName,
		},
		Spec: k8sadmission.MutatingAdmissionPolicySpec{
			// Do not block creation of the route controller in case of failure
			FailurePolicy: ptr.To(k8sadmission.Ignore),
			// Only run once
			ReinvocationPolicy: k8sadmission.NeverReinvocationPolicy,
			MatchConstraints: &k8sadmission.MatchResources{
				// Scope to the shoot's control-plane namespace only
				NamespaceSelector: &k8smeta.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": cpNS,
					},
				},
				// Scope to route-controller pods by their own label selector
				ObjectSelector: &k8smeta.LabelSelector{
					MatchLabels: rcLabels,
				},
				// Match RC pod creation requests
				ResourceRules: []k8sadmission.NamedRuleWithOperations{
					{
						RuleWithOperations: k8sadmission.RuleWithOperations{
							Operations: []k8sadmission.OperationType{k8sadmission.Create},
							Rule: k8sadmission.Rule{
								APIGroups:   []string{""},
								APIVersions: []string{"v1"},
								Resources:   []string{"pods"},
							},
						},
					},
				},
			},
			// Skip if the init container is already present (idempotent).
			MatchConditions: []k8sadmission.MatchCondition{
				{
					Name: "skip-if-already-delayed",
					Expression: fmt.Sprintf(
						`!has(object.spec.initContainers) || !object.spec.initContainers.exists(c, c.name == %q)`,
						routeDelayInitContainer),
				},
			},
			Mutations: []k8sadmission.Mutation{
				{
					PatchType: k8sadmission.PatchTypeJSONPatch,
					JSONPatch: &k8sadmission.JSONPatch{
						Expression: fmt.Sprintf(
							`has(object.spec.initContainers) ?
								[JSONPatch{op: "add", path: "/spec/initContainers/-",
									value: {"name": dyn(%q), "image": dyn(%q), "command": dyn(["sh", "-c", "sleep %d"])}}]
								:
								[JSONPatch{op: "add", path: "/spec/initContainers",
									value: dyn([{"name": dyn(%q), "image": dyn(%q), "command": dyn(["sh", "-c", "sleep %d"])}])}]`,
							routeDelayInitContainer, busyboxImage, int(routeDelay.Seconds()),
							routeDelayInitContainer, busyboxImage, int(routeDelay.Seconds()),
						),
					},
				},
			},
		},
	}
	Expect(createOnSeedServedVersion(ctx, f, mapObj, "MutatingAdmissionPolicy")).To(Succeed(),
		"failed to create MutatingAdmissionPolicy %q on seed", mapPolicyName)

	// Defer: release the hazard (delete MAP + Binding AND clear the sleeping route-controller pod).
	// Registered here (after MAP creation, before Binding) so cleanup fires even if Binding creation
	// fails. Clearing the pod — not just the MAP — matters on the failure path: the injected sleep
	// otherwise keeps routes blocked (and the shoot's overlay switch wedged) until it expires.
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		releaseRouteDelay(cleanCtx, f, cpNS, rcLabels, mapPolicyName, mapBindingName)
	}()

	mapBindingObj := &k8sadmission.MutatingAdmissionPolicyBinding{
		ObjectMeta: k8smeta.ObjectMeta{
			Name: mapBindingName,
		},
		Spec: k8sadmission.MutatingAdmissionPolicyBindingSpec{
			PolicyName: mapPolicyName,
		},
	}
	Expect(createOnSeedServedVersion(ctx, f, mapBindingObj, "MutatingAdmissionPolicyBinding")).To(Succeed(),
		"failed to create MutatingAdmissionPolicyBinding %q on seed", mapBindingName)

	// ── Step 7: delete existing route-controller pod ─────────────────────────────
	// The route controller runs continuously; Gardener does not restart it when
	// switching overlay modes. Deleting the pod forces the ReplicaSet to recreate
	// it, giving the MAP a CREATE request to intercept.
	By("Delete existing route-controller pod to trigger MAP-intercepted recreation")
	rcPods := &k8score.PodList{}
	Expect(f.SeedClient.Client().List(ctx, rcPods,
		client.InNamespace(cpNS),
		client.MatchingLabels(rcLabels),
	)).To(Succeed())
	for _, pod := range rcPods.Items {
		Expect(f.SeedClient.Client().Delete(ctx, &pod)).To(Succeed())
		f.Logger.Info("Deleted route-controller pod to trigger MAP-intercepted recreation", "pod", pod.Name)
	}

	// ── Step 8: flip overlay off ──────────────────────────────────────────────────
	// Record the moment the switch is triggered; only probe FAILs at or after this time
	// count towards the zero-FAIL assertion (step 13), so pre-switch startup noise is ignored.
	switchStartUnix := time.Now().Unix()
	By("Switch: set overlay=false in shoot providerConfig")
	netConfig := &gextcalico.NetworkConfig{
		TypeMeta: k8smeta.TypeMeta{
			APIVersion: "calico.networking.extensions.gardener.cloud/v1alpha1",
			Kind:       "NetworkConfig",
		},
		Overlay: &gextcalico.Overlay{Enabled: false},
	}
	netConfigJSON, err := json.Marshal(netConfig)
	Expect(err).NotTo(HaveOccurred())

	Expect(f.UpdateShootSpec(ctx, f.Shoot, func(shoot *gcore.Shoot) error {
		if shoot.Spec.Networking == nil {
			shoot.Spec.Networking = &gcore.Networking{}
		}
		shoot.Spec.Networking.ProviderConfig = &runtime.RawExtension{Raw: netConfigJSON}
		return nil
	})).To(Succeed())

	By("Annotate shoot with gardener.cloud/operation=reconcile")
	Expect(f.AnnotateShoot(ctx, f.Shoot, map[string]string{
		gcoreconst.GardenerOperation: gcoreconst.GardenerOperationReconcile,
	})).To(Succeed())

	// ── Step 9: wait for the trigger to engage and the gate to hold ───────────────
	By(fmt.Sprintf("Wait for route-controller pod with init container %q to appear on seed (up to %s)",
		routeDelayInitContainer, triggerEngagedTimeout))

	Eventually(func(g Gomega) bool {
		podList := &k8score.PodList{}
		g.Expect(f.SeedClient.Client().List(ctx, podList,
			client.InNamespace(cpNS),
			client.MatchingLabels(rcLabels),
		)).To(Succeed())

		for _, pod := range podList.Items {
			for _, ic := range pod.Status.InitContainerStatuses {
				if ic.Name == routeDelayInitContainer {
					f.Logger.Info("Trigger engaged: route-controller pod is running sleep init container",
						"pod", pod.Name, "initContainer", ic.Name, "state", ic.State)
					return true
				}
			}
		}
		return false
	}).WithContext(ctx).WithTimeout(triggerEngagedTimeout).WithPolling(triggerPollInterval).Should(BeTrue(),
		"route-controller pod with init container %q did not appear in %q within %s; "+
			"possible causes: MAP not admitted (check seed API server) "+
			"or route-controller Deployment labels changed",
		routeDelayInitContainer, cpNS, triggerEngagedTimeout)

	// Wait for the extension reconcile to actually reach the overlay-switch gate and hold.
	// The gate records gateHoldErrorSubstring on the Network's status.lastError while routes
	// are absent. Synchronising on this signal — not merely on the sleep init container's
	// presence — is what makes the gate-hold assertion below meaningful: it proves the
	// reconcile reached the networking step and is holding overlay because routes are missing.
	By(fmt.Sprintf("Wait for networking-calico reconcile to reach the overlay-switch gate "+
		"(Network status.lastError contains %q, up to %s)", gateHoldErrorSubstring, gateEngagedTimeout))
	Eventually(func(g Gomega) string {
		nwList := &unstructured.UnstructuredList{}
		nwList.SetGroupVersionKind(networkListGVK)
		g.Expect(f.SeedClient.Client().List(ctx, nwList, client.InNamespace(cpNS))).To(Succeed())
		g.Expect(nwList.Items).NotTo(BeEmpty(), "no Network resource found in %q", cpNS)
		desc, _, _ := unstructured.NestedString(nwList.Items[0].Object, "status", "lastError", "description")
		return desc
	}).WithContext(ctx).WithTimeout(gateEngagedTimeout).WithPolling(gateEngagedPollInterval).
		Should(ContainSubstring(gateHoldErrorSubstring),
			"networking-calico reconcile did not reach the overlay-switch gate within %s; "+
				"the gate-hold assertion would be vacuous without it (routes may have been created "+
				"before the reconcile reached the networking step — consider raising routeDelay)",
			gateEngagedTimeout)

	// ── Step 10: assert the gate holds overlay while routes are absent ────────────
	By(fmt.Sprintf("Assert gate holds overlay for %s while routes are absent", gateAssertWindow))
	observed := false
	Consistently(func(g Gomega) {
		// Transient API errors are tolerated (skip the iteration): they are not evidence the
		// gate flipped, and failing on them would make this window assertion flaky.
		mode, getErr := getIPPoolIPIPMode(ctx, shootClient.Client())
		if getErr != nil {
			f.Logger.Info("IPPool get failed during gate-hold check (transient?)", "error", getErr)
			return
		}
		nodes := &k8score.NodeList{}
		if listErr := shootClient.Client().List(ctx, nodes); listErr != nil {
			f.Logger.Info("Node list failed during gate-hold check (transient?)", "error", listErr)
			return
		}
		observed = true

		g.Expect(mode).To(Equal("Always"),
			"SeamlessOverlaySwitch gate FAILED: IPPool ipipMode flipped to %q while routes are "+
				"still absent (route-controller sleeping)", mode)
		g.Expect(checkNodesNetworkAvailable(nodes, "")).To(Succeed())
		// Confirm the hazard is genuinely present: if every node already reports RouteCreated
		// the routes exist and the window proves nothing.
		g.Expect(nodesAllRouteCreated(nodes)).To(BeFalse(),
			"gate-hold window is vacuous: all nodes already report RouteCreated (routes exist, "+
				"hazard not active)")
	}).WithContext(ctx).WithTimeout(gateAssertWindow).WithPolling(gateAssertInterval).Should(Succeed())
	Expect(observed).To(BeTrue(),
		"gate-hold window elapsed without a single successful IPPool+node observation")

	// ── Step 11: release the hazard and wait for the switch to complete ───────────
	By("Release hazard: delete MAP and ensure a clean route-controller pod (no sleep) is running")
	releaseRouteDelay(ctx, f, cpNS, rcLabels, mapPolicyName, mapBindingName)

	By("Wait for shoot to be fully reconciled (routes created → gate unblocks → switch completes)")
	Expect(f.WaitForShootToBeReconciled(ctx, f.Shoot)).To(Succeed())

	// ── Step 12: post-switch assertions ──────────────────────────────────────────
	By("Assert post-switch: IPPool ipipMode=Never")
	finalMode, err := getIPPoolIPIPMode(ctx, shootClient.Client())
	Expect(err).NotTo(HaveOccurred())
	Expect(finalMode).To(Equal("Never"),
		"after switch, IPPool ipipMode must be Never; got %q", finalMode)

	By("Assert post-switch: all nodes NetworkUnavailable status=False reason=RouteCreated")
	postNodes := &k8score.NodeList{}
	Expect(shootClient.Client().List(ctx, postNodes)).To(Succeed())
	Expect(checkNodesNetworkAvailable(postNodes, "RouteCreated")).To(Succeed())

	// ── Step 13: collect probe logs + assert zero cross-node FAIL ─────────────────
	By("Collect probe client logs and assert zero cross-node FAIL after the switch")
	clientPods := &k8score.PodList{}
	Expect(shootClient.Client().List(ctx, clientPods,
		client.InNamespace(probeNamespace),
		client.MatchingLabels{"app": "probe-client"},
	)).To(Succeed())

	crossNodeFail, sameNodeFail := 0, 0
	var failLines []string
	for _, pod := range clientPods.Items {
		logBytes, logErr := kubernetesutils.GetPodLogs(ctx,
			shootClient.Kubernetes().CoreV1().Pods(probeNamespace),
			pod.Name,
			&k8score.PodLogOptions{},
		)
		if logErr != nil {
			f.Logger.Info("failed to collect probe client logs (continuing)",
				"pod", pod.Name, "error", logErr)
			continue
		}
		// Each line: "<unix_ts> src=<node> dst=<ip> OK|FAIL"
		for line := range strings.SplitSeq(string(logBytes), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 4 || fields[len(fields)-1] != "FAIL" {
				continue
			}
			// Only count FAILs at or after the switch was triggered; earlier lines are
			// probe-mesh startup noise, not the switch window under test.
			ts, parseErr := strconv.ParseInt(fields[0], 10, 64)
			if parseErr != nil || ts < switchStartUnix {
				continue
			}
			srcNode := strings.TrimPrefix(fields[1], "src=")
			dstIP := strings.TrimPrefix(fields[2], "dst=")
			// Classify: a FAIL to a server pod on the same node is a control flow that should
			// never break regardless of the switch; cross-node is the property under test.
			if node, ok := serverIPToNode[dstIP]; ok && node == srcNode {
				sameNodeFail++
			} else {
				crossNodeFail++
			}
			if len(failLines) < 100 { // cap output to avoid noise
				failLines = append(failLines, line)
			}
		}
	}

	if len(failLines) > 0 {
		GinkgoWriter.Printf("Probe FAIL lines after switch (first %d; cross-node=%d same-node=%d):\n%s\n",
			len(failLines), crossNodeFail, sameNodeFail, strings.Join(failLines, "\n"))
	}
	if sameNodeFail > 0 {
		// Same-node flows breaking points at environmental flakiness (node/pod churn, probe
		// timeouts under load), not the gate. Surface it, but do not fail the gate assertion on it.
		f.Logger.Info("WARNING: same-node (control) probes failed after the switch; cross-node "+
			"results may be influenced by environmental flakiness unrelated to the gate",
			"sameNodeFail", sameNodeFail)
	}

	Expect(crossNodeFail).To(BeZero(),
		"expected zero cross-node FAIL events after the switch was triggered; "+
			"SeamlessOverlaySwitch gate should have held overlay until routes existed — "+
			"if cross-node FAIL > 0 the gate did not protect the switch window")
}

// admissionRegistrationGroup is the API group serving MutatingAdmissionPolicy and its Binding.
const admissionRegistrationGroup = "admissionregistration.k8s.io"

// seedServedGVK resolves, via the seed's RESTMapper (server discovery), the GroupVersionKind at
// which the seed actually serves the given admissionregistration kind. MutatingAdmissionPolicy is
// served at v1alpha1 (k8s 1.32/1.33, and on 1.34/1.35 when the beta is not opted in), v1beta1
// (1.34/1.35), or v1 (1.36+ GA). All three share identical JSON for the fields this test sets, so
// the test authors the object once with the typed v1 package, converts it to unstructured, and
// stamps whichever version the seed reports.
func seedServedGVK(f *framework.ShootFramework, kind string) (schema.GroupVersionKind, error) {
	gk := schema.GroupKind{Group: admissionRegistrationGroup, Kind: kind}
	mapping, err := f.SeedClient.Client().RESTMapper().RESTMapping(gk)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("seed does not serve %s.%s (needs v1alpha1, v1beta1 or v1): %w", kind, admissionRegistrationGroup, err)
	}
	return mapping.GroupVersionKind, nil
}

// createOnSeedServedVersion creates a typed admissionregistration object on the seed at the
// version the seed serves for the given kind. It converts the typed object to unstructured and
// stamps the discovered GVK, so it is agnostic to the seed's k8s version (v1alpha1/v1beta1/v1,
// whose schemas are identical for the fields set here) and to which versions the client scheme has.
func createOnSeedServedVersion(ctx context.Context, f *framework.ShootFramework, obj client.Object, kind string) error {
	gvk, err := seedServedGVK(f, kind)
	if err != nil {
		return err
	}
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return fmt.Errorf("converting %s to unstructured: %w", kind, err)
	}
	u := &unstructured.Unstructured{Object: raw}
	u.SetGroupVersionKind(gvk)
	return f.SeedClient.Client().Create(ctx, u)
}

// routeDelayMAPNames returns the per-shoot MutatingAdmissionPolicy and Binding names,
// suffixing the base names with a short hash of the control-plane namespace. This keeps
// concurrent runs against different shoots on the same seed from sharing (and deleting)
// each other's cluster-scoped objects, while a re-run of the same shoot reuses the same
// names so the pre-create cleanup and defer stay idempotent.
func routeDelayMAPNames(cpNS string) (policy, binding string) {
	sum := sha256.Sum256([]byte(cpNS))
	suffix := hex.EncodeToString(sum[:])[:8]
	return mapPolicyBaseName + "-" + suffix, mapBindingBaseName + "-" + suffix
}

// deleteRouteDelayMAP deletes the given MutatingAdmissionPolicy and Binding on the seed.
// It is idempotent: IgnoreNotFound is applied so it is safe to call from a defer even
// when the objects were never created or were already deleted.
func deleteRouteDelayMAP(ctx context.Context, f *framework.ShootFramework, policyName, bindingName string) {
	for _, res := range []struct{ kind, name string }{
		{"MutatingAdmissionPolicy", policyName},
		{"MutatingAdmissionPolicyBinding", bindingName},
	} {
		gvk, err := seedServedGVK(f, res.kind)
		if err != nil {
			f.Logger.Error(err, "cleanup: cannot resolve served version", "kind", res.kind, "name", res.name)
			continue
		}
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		obj.SetName(res.name)
		if err := f.SeedClient.Client().Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
			f.Logger.Error(err, "cleanup: failed to delete", "kind", res.kind, "name", res.name)
		}
	}
}

// releaseRouteDelay removes the MAP/Binding and ensures a route-controller pod is running WITHOUT
// the injected sleep, so VPC routes get created and the SeamlessOverlaySwitch gate releases.
//
// It guards the admission race on teardown: deleting the MAP and the pod in quick succession can
// let the ReplicaSet recreate the pod before the policy deletion has propagated to the seed
// apiserver, which would re-inject the sleep. So after deleting the MAP it keeps deleting any pod
// that still carries the sleep init container until a clean one appears. Best-effort (logs, never
// fails) so it is safe to call from a cleanup defer as well as the happy path.
func releaseRouteDelay(ctx context.Context, f *framework.ShootFramework, cpNS string, rcLabels map[string]string, policyName, bindingName string) {
	deleteRouteDelayMAP(ctx, f, policyName, bindingName)

	if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 3*time.Minute, true,
		func(ctx context.Context) (bool, error) {
			pods := &k8score.PodList{}
			if err := f.SeedClient.Client().List(ctx, pods,
				client.InNamespace(cpNS), client.MatchingLabels(rcLabels)); err != nil {
				f.Logger.Info("release: listing route-controller pods failed (retrying)", "error", err)
				return false, nil
			}
			if len(pods.Items) == 0 {
				return false, nil // wait for the ReplicaSet to recreate one
			}
			stillSleeping := false
			for i := range pods.Items {
				pod := &pods.Items[i]
				if !hasRouteDelayInitContainer(pod) {
					continue
				}
				stillSleeping = true
				if err := f.SeedClient.Client().Delete(ctx, pod); client.IgnoreNotFound(err) != nil {
					f.Logger.Error(err, "release: failed to delete sleeping route-controller pod", "pod", pod.Name)
				} else {
					f.Logger.Info("release: deleted sleeping route-controller pod", "pod", pod.Name)
				}
			}
			return !stillSleeping, nil // done once no pod carries the sleep init container
		}); err != nil {
		f.Logger.Error(err, "release: a route-controller pod still carried the sleep init container "+
			"after the deadline; routes may stay blocked until the injected sleep expires")
	}
}

// hasRouteDelayInitContainer reports whether the pod carries the injected route-delay init container.
func hasRouteDelayInitContainer(pod *k8score.Pod) bool {
	for _, c := range pod.Spec.InitContainers {
		if c.Name == routeDelayInitContainer {
			return true
		}
	}
	return false
}

// getIPPoolIPIPMode reads spec.ipipMode from the Calico default IPPool on the given cluster.
func getIPPoolIPIPMode(ctx context.Context, c client.Client) (string, error) {
	ippool := &unstructured.Unstructured{}
	ippool.SetGroupVersionKind(ippoolGVK)
	if err := c.Get(ctx, client.ObjectKey{Name: ippoolName}, ippool); err != nil {
		return "", err
	}
	mode, found, err := unstructured.NestedString(ippool.Object, "spec", "ipipMode")
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("IPPool %q spec.ipipMode field not found", ippoolName)
	}
	return mode, nil
}

// checkNodesNetworkAvailable verifies every node's NodeNetworkUnavailable condition reports
// the network as available (status=False). If requireReason is non-empty, the condition
// reason must also match. Returns an error describing the first violation, or nil if all pass.
func checkNodesNetworkAvailable(nodes *k8score.NodeList, requireReason string) error {
	for _, node := range nodes.Items {
		for _, cond := range node.Status.Conditions {
			if cond.Type != k8score.NodeNetworkUnavailable {
				continue
			}
			if cond.Status != k8score.ConditionFalse {
				return fmt.Errorf("node %q NetworkUnavailable status must be False (network available); got %q",
					node.Name, cond.Status)
			}
			if requireReason != "" && cond.Reason != requireReason {
				return fmt.Errorf("node %q NetworkUnavailable reason must be %q; got %q",
					node.Name, requireReason, cond.Reason)
			}
		}
	}
	return nil
}

// nodesAllRouteCreated reports whether every node already has its NodeNetworkUnavailable
// condition set to False with reason RouteCreated (i.e. VPC routes exist on all nodes).
// Returns false for an empty list.
func nodesAllRouteCreated(nodes *k8score.NodeList) bool {
	if len(nodes.Items) == 0 {
		return false
	}
	for _, node := range nodes.Items {
		routeCreated := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == k8score.NodeNetworkUnavailable &&
				cond.Status == k8score.ConditionFalse && cond.Reason == "RouteCreated" {
				routeCreated = true
			}
		}
		if !routeCreated {
			return false
		}
	}
	return true
}
