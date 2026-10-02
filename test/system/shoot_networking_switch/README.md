# Calico Overlay-to-Native Seamless Switch — System Test

## Purpose

This package is a Gardener system/integration test that verifies the Calico `SeamlessOverlaySwitch`
feature gate prevents cross-node pod-connectivity loss when a shoot is switched from IP-in-IP overlay
mode (`ipipMode=Always`) to native routing (`ipipMode=Never`).

Without the gate, patching `overlay.enabled=false` causes Calico to reconfigure the IPPool
immediately, dropping cross-node traffic until VPC pod-CIDR routes exist. The gate holds the IPPool
in overlay mode until the AWS route controller has created those routes — signalled by
`NetworkUnavailable` reason `RouteCreated` on every node.

**Why this test requires a real cloud shoot.** The test looks for the `aws-custom-route-controller`
Deployment on the seed. This component is only deployed on AWS-backed seeds. It is the one that sets
the `RouteCreated` condition; without it there is no signal for the gate to wait for. KinD /
provider-local environments have no route controller and no VPC routes, so the Deployment lookup
fails the spec (step 5) and the gate invariant can never be satisfied.

The test fails fast (Ginkgo `Fail`) if the `SeamlessOverlaySwitch` feature gate is not enabled on
the `networking-calico` ControllerDeployment in the garden cluster (step 1).

---

## How the Test Works

`RunTest` in `networking_switch_test.go` proceeds in thirteen numbered steps.

### Step 1 — Gate precondition

The test reads the `networking-calico` ControllerDeployment from the garden cluster and checks
`helm.values.config.featureGates.SeamlessOverlaySwitch`. If the field is absent or `false`, the spec
fails fast with `Fail(...)` — the gate must be enabled for the test to be meaningful.

### Step 2 — Shoot client and start-state assertion

A shoot client is created from the admin kubeconfig, then one invariant is verified before anything
is changed:

- The `default-ipv4-ippool` Calico IPPool on the shoot must have `spec.ipipMode=Always`.

This catches a non-overlay shoot before the test makes any changes. (Node `NetworkUnavailable`
conditions are only asserted after the switch, in step 12; the start state is assumed valid per the
manual-run prerequisites.)

### Step 3 — Deploy the probe mesh

The template `templates/network-probe-mesh.yaml.tpl` is rendered in two passes into the
`calico-switch-test` namespace on the shoot cluster.

**Pass 1 — server only.** The `probe-server` DaemonSet is deployed with `Targets=""` (pod IPs are
not yet known). It runs one replica per node using `agnhost:2.47 netexec --http-port=8080`.

After `WaitUntilDaemonSetIsRunning`, the test lists all `probe-server` pods and collects their IPs.

**Pass 2 — full mesh.** The template is re-rendered with the collected IPs as the `Targets` value
(space-separated). The `probe-client` DaemonSet (`busybox:1.36`) starts. Each client pod runs a
tight loop firing concurrent `wget` requests every `0.3s` to every server IP and logging:

```
<unix_ts> src=<nodeName> dst=<podIP> OK|FAIL
```

Cross-node probes are the flows that break during an unprotected switch; same-node probes serve as a
control. Because pods write locally throughout, no events are missed even if the shoot VPN is
temporarily degraded mid-switch. Log collection is deferred to step 13, after reconcile completes.

A `defer` deletes the `calico-switch-test` namespace on test exit (both success and failure).

### Step 4 — Prepare the seed client

The seed name is read from `shoot.spec.seedName`. The test calls `f.GetSeed` to obtain a seed
client. The control-plane namespace is taken from `shoot.status.technicalID`.

### Step 5 — Discover the route-controller pod selector

The `aws-custom-route-controller` Deployment is fetched from the control-plane namespace on the
seed; if it is absent (non-AWS seed) the spec fails here. Its `spec.selector.matchLabels` are
extracted. These labels are used to scope the MAP (step 6), to delete the running pod (step 7), and
to poll for the recreated pod (step 9).

### Step 6 — Install the MAP trigger

A `MutatingAdmissionPolicy` named `calico-switch-test-route-delay-<hash>` plus a
`MutatingAdmissionPolicyBinding` named `calico-switch-test-route-delay-binding-<hash>` are created on
the **seed** cluster (`<hash>` is a short hash of the shoot's control-plane namespace, so concurrent
runs against different shoots on the same seed get distinct objects and don't delete each other's).
Together they are the "trigger" — the mechanism that deterministically creates the
"routes not yet created" hazard window.

When a route-controller pod is created, the MAP intercepts the new Pod's `CREATE` admission request
and appends an init container via JSONPatch:

```
op: add, path: /spec/initContainers/-
value: {name: "route-creation-delay", image: "busybox:1.36",
        command: ["sh", "-c", "sleep <routeDelaySeconds>"]}
```

The pod blocks in the init phase for `routeDelaySeconds` seconds. The route controller binary never
starts, so VPC routes are not created — exactly the hazard the gate must handle.

**Why the trigger is surgical.** The MAP uses three layers of scoping:
- `namespaceSelector` pins to the exact control-plane namespace (`kubernetes.io/metadata.name: <cpNS>`).
- `objectSelector` uses the labels discovered from the Deployment, so only route-controller pods match.
- `resourceRules` restricts to `pods` / `CREATE` only.

Targeting the Pod (not the Deployment) means gardener-resource-manager, which owns the Deployment,
never sees the mutation and does not fight it.

**Why the trigger is self-healing.** The sleep expires on its own. Even if the test crashes before the
cleanup `defer` fires, the sleep finishes, the route controller starts, VPC routes are created, and
the shoot converges normally. The `defer` deletes the MAP and Binding for hygiene, but connectivity
recovery does not depend on it.

A `matchCondition` CEL expression guards idempotency:
```
!has(object.spec.initContainers) ||
!object.spec.initContainers.exists(c, c.name == "route-creation-delay")
```
If the init container is already present (e.g. on MAP re-invocation), the patch is skipped.
`failurePolicy: Ignore` prevents a CEL error from blocking pod scheduling.
`reinvocationPolicy: Never` prevents redundant re-runs.

### Step 7 — Delete the running route-controller pod

The `aws-custom-route-controller` pod runs continuously; Gardener does **not** restart it when
overlay mode is switched. The test deletes the running pod so the ReplicaSet immediately recreates
it, giving the MAP a Pod `CREATE` to intercept. This — not the reconcile — is what injects the sleep
init container, so the hazard window is armed deterministically rather than depending on reconcile
timing.

### Step 8 — Flip overlay off

`shoot.spec.networking.providerConfig` is patched to set `overlay.enabled=false`, serialised as a
`NetworkConfig` (`calico.networking.extensions.gardener.cloud/v1alpha1`). The shoot is immediately
annotated with `gardener.cloud/operation=reconcile` to trigger Gardener processing.

### Step 9 — Wait for the trigger to engage and the gate to hold

Two waits, in order:

1. **Trigger armed.** The test polls the seed for route-controller pods (labels from step 5) and
   inspects `status.initContainerStatuses` until a pod lists `route-creation-delay` — i.e. the MAP
   intercepted the Pod recreated in step 7. Up to `triggerEngagedTimeout` at `triggerPollInterval`.
2. **Gate holding.** The test then polls the `Network` extension resource in the control-plane
   namespace on the seed until its `status.lastError.description` contains *"waiting for routes to be
   created"* — the error the networking-calico actuator records while it holds the switch
   ([actuator_reconcile.go](../../../pkg/controller/actuator_reconcile.go)). Up to
   `gateEngagedTimeout` at `gateEngagedPollInterval`.

The second wait is the key correctness point: it proves the reconcile actually reached the networking
step, detected the switch, and is holding overlay *because routes are absent* — so the assertion in
step 10 is not vacuous. If the reconcile never reaches the gate within `gateEngagedTimeout`
(e.g. routes were created before gardenlet got to the networking step), the spec fails here rather
than passing a meaningless window.

### Step 10 — Assert the gate holds

With the gate confirmed holding, the test enters a `Consistently` loop for `gateAssertWindow`
(sampled at `gateAssertInterval`). On each iteration it checks:

- `default-ipv4-ippool` still has `spec.ipipMode=Always`.
- Every node's `NetworkUnavailable` condition is still `status=False` (transient API errors are
  tolerated — the iteration is skipped, not failed).
- Routes are still absent (not every node reports `reason=RouteCreated`), confirming the hazard is
  genuinely active during the window.

It also requires at least one fully successful observation, so a window spent entirely on API errors
cannot pass. Any flip fails the spec immediately. `gateAssertWindow` is shorter than `routeDelay`, so
the hazard is still active for the entire window.

> **Note.** The IPPool-hold assertion is now synchronized with the gate (step 9), so it is meaningful
> rather than timing-dependent. The end-to-end probe check in step 13 remains the authoritative
> guarantee that no cross-node traffic was dropped.

### Step 11 — Release the hazard and complete the switch

The test deletes the MAP + Binding and then ensures a route-controller pod runs **without** the sleep
init container, so VPC routes get created and the gate releases — rather than waiting out the full
`routeDelay` sleep. Because deleting the MAP and the pod in quick succession races the policy's
admission teardown (a pod recreated too soon can get the sleep re-injected), the release keeps
deleting any pod that still carries the `route-creation-delay` init container until a clean one
appears. The test then calls `f.WaitForShootToBeReconciled`: routes created → nodes transition to
`RouteCreated` → the gate unblocks → Calico applies `ipipMode=Never` → Gardener reports `Succeeded`.

The same release runs from the teardown `defer`, so a run that fails *before* this step still clears
the sleeping pod and does not leave the shoot's overlay switch wedged until the sleep expires.

### Step 12 — Post-switch assertions

- `default-ipv4-ippool` must have `spec.ipipMode=Never`.
- Every node's `NetworkUnavailable` condition must have `status=False` and `reason=RouteCreated`.

### Step 13 — Probe log collection and zero-FAIL assert

Logs are fetched from every `probe-client` pod. Each `FAIL` line *at or after the switch was
triggered* is classified by comparing the destination server pod's node to the source node:
same-node FAILs (control flows) are logged as a flakiness warning; cross-node FAILs are the property
under test. The spec requires zero cross-node FAILs — if any occur, the gate did not hold overlay
until routes existed and cross-node traffic was disrupted.

---

## MutatingAdmissionPolicy

This test uses a `MutatingAdmissionPolicy` (MAP) on the seed as the trigger (see Step 6) — an
in-tree, CEL-based admission mutation evaluated inside the kube-apiserver. For background on MAP
and its Policy/Binding object pair, see the Kubernetes documentation:
<https://kubernetes.io/docs/reference/access-authn-authz/mutating-admission-policy/>.

MAP is served at `v1alpha1` (k8s 1.32/1.33, and 1.34/1.35 without the beta opt-in), `v1beta1`
(1.34/1.35), or `v1` (1.36+ GA). These share identical JSON for the fields this test sets, so the
test authors the object once and submits it at whichever version the **seed** serves (discovered via
the seed's RESTMapper). The only seed requirement is that the MAP API is served at all — any of the
three versions works.

### How this test uses MAP

The policy and binding (`calico-switch-test-route-delay-<hash>` / `…-binding-<hash>`) are
created on the **seed** cluster (not the shoot cluster). The seed's kube-apiserver evaluates the
policy when a new `aws-custom-route-controller` Pod is admitted.

The JSONPatch mutation expression is:

```
[JSONPatch{
  op: "add",
  path: "/spec/initContainers/-",
  value: {
    "name": "route-creation-delay",
    "image": "busybox:1.36",
    "command": ["sh", "-c", "sleep <routeDelaySeconds>"]
  }
}]
```

The `/spec/initContainers/-` JSON Pointer appends to the array. Kubernetes always initialises
`spec.initContainers` to `[]` in the admission payload, so the append pointer is safe even when no
prior init containers exist.

The `matchCondition` CEL expression makes the mutation idempotent: if `route-creation-delay` is
already in `initContainers` (possible if the MAP is re-invoked), the patch is skipped.

---

## Ginkgo Integration

### Suite wiring

Ginkgo requires a standard Go test function as its entry point. `networking_switch_suite_test.go`
provides exactly this and nothing more:

**`init()`** calls `framework.RegisterShootFrameworkFlags()`. This registers Gardener test framework
flags (e.g. `-kubecfg`, `-shoot-name`, `-project-namespace`) as standard `flag` variables before
`flag.Parse()` is called. Because `init()` runs before the test binary parses its arguments, the
flags are available to the Ginkgo runner when it starts.

**`TestNetworkingSwitch(t *testing.T)`** is the entry point:
1. `RegisterFailHandler(Fail)` — wires Gomega assertion failures into Ginkgo's failure path.
2. `RunSpecs(t, "Calico Overlay-Native Seamless Switch Test Suite")` — hands control to the Ginkgo
   runner, which discovers and runs all `Describe`/`It` nodes registered via `var _ = Describe(...)`.

### Spec declaration

`networking_switch_test.go` registers one spec:

```go
var _ = Describe("Calico overlay->native seamless switch", func() {
    f := framework.NewShootFramework(nil)
    framework.CIt("should switch overlay->native with zero connectivity loss", func(ctx context.Context) {
        RunTest(ctx, f)
    }, SwitchTestTimeout)
})
```

`framework.CIt` is a Gardener helper that wraps `It` with a context-aware deadline (here
`SwitchTestTimeout`). The context `ctx` passed to the test function is cancelled when the deadline
expires, causing any in-flight Kubernetes API calls to return promptly.

`framework.NewShootFramework` creates a `ShootFramework` that wires up garden, seed, and shoot
clients in its `BeforeEach`. On spec failure, the framework dumps shoot and seed state to the Ginkgo
report — crucial for post-mortem debugging.

### Passing flags to `go test`

The flags registered by `RegisterShootFrameworkFlags()` are passed as standard `go test` flag
arguments:

```
go test ... -kubecfg=<path> -shoot-name=<name> -project-namespace=garden-<project>
```

Ginkgo's own flags use the `-ginkgo.` prefix (e.g. `-ginkgo.v` for verbose output).

### TestMachinery

The suite runs both standalone (manual `go test` invocation) and under Gardener TestMachinery. The
TestDefinition at `.test-defs/SeamlessOverlaySwitchTest.yaml` invokes the same `go test` binary with
flags injected from environment variables, with an `activeDeadlineSeconds` matching the spec timeout.

---

## How to Run Manually

### Prerequisites

1. The target shoot is in overlay mode:
   - `spec.networking.providerConfig` has `overlay.enabled=true`.
   - The Calico `default-ipv4-ippool` IPPool has `spec.ipipMode=Always`.
   - All nodes have `NetworkUnavailable` with `reason=CalicoIsUp`.
   - VPC pod-CIDR routes are absent (the route controller has not yet created them).
2. The `SeamlessOverlaySwitch` feature gate is enabled on the `networking-calico`
   ControllerDeployment in the garden cluster:
   `helm.values.config.featureGates.SeamlessOverlaySwitch: true`.
3. The shoot uses an AWS seed — the `aws-custom-route-controller` Deployment must be present in
   the shoot's control-plane namespace on the seed.
4. Your garden kubeconfig has read access to the garden cluster (ControllerDeployment, Shoot) and
   create/delete access on the seed cluster (MutatingAdmissionPolicy, MutatingAdmissionPolicyBinding).

### Command

From the repository root:

```sh
go test -timeout=0 \
  ./test/system/shoot_networking_switch/... \
  -kubecfg=<path-to-garden-kubeconfig> \
  -shoot-name=<shoot-name> \
  -project-namespace=garden-<project-name> \
  -ginkgo.v
```

`-timeout=0` disables Go's default 10-minute test timeout. The Ginkgo spec enforces its own
deadline via `SwitchTestTimeout`.

### What to expect

The dominant phases of wall-clock time are:

- Waiting for Gardener to reconcile far enough to reach the networking step and hit the overlay-switch
  gate, observed via the `Network` resource's `status.lastError` (up to `gateEngagedTimeout`). This is
  usually the longest phase and varies by landscape.
- The gate-hold assert window (`gateAssertWindow`), which runs while routes are still absent.
- The final reconcile after the hazard is released, which completes once VPC routes are created.

The spec fails fast if a precondition is violated, if the reconcile never reaches the gate, if the
gate does not hold overlay while routes are absent, or if any cross-node `FAIL` lines appear in the
probe client logs after the switch.
