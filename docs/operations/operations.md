# Using the Calico networking extension with Gardener as operator

This document explains configuration options supported by the networking-calico extension.

### Run calico-node in non-privileged and non-root mode

**Feature State**: `Alpha`

##### Motivation

Running containers in privileged mode is not recommended as privileged containers run with all [linux capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html) enabled and can access the host's resources. Running containers in privileged mode opens number of security threats such as breakout to underlying host OS.

##### Support for non-privileged and non-root mode

The Calico project has a preliminary support for running the calico-node component in non-privileged mode. Similar to [Tigera Calico operator](https://github.com/tigera/operator) the networking-calico extension can also run calico-node in non-privileged and non-root mode. This feature is controller via feature gate named `NonPrivilegedCalicoNode`. The feature gates are configured in the [ControllerConfiguration](../../example/00-componentconfig.yaml) of networking-calico. The corresponding ControllerDeployment configuration that enables the `NonPrivilegedCalicoNode` would look like:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: ControllerDeployment
metadata:
  name: networking-calico
type: helm
providerConfig:
  values:
    chart: <omitted>
    config:
      featureGates:
        NonPrivilegedCalicoNode: false
```

##### Limitations

- The support for the non-privileged mode in the Calico project is not ready for productive usage. The [upstream documentation](https://projectcalico.docs.tigera.io/security/non-privileged) states that in non-privileged mode the support for features added after Calico v3.21 is not guaranteed.
- Calico in non-privileged mode does not support eBPF dataplane. That's why when eBPF dataplane is enabled, calico-node has to run in privileged mode (even when the `NonPrivilegedCalicoNode` feature gate is enabled).
- (At the time of writing this guide) there is the following issue [projectcalico/calico#5348](https://github.com/projectcalico/calico/issues/5348) that is not addressed.
- (At the time of writing this guide) the upstream adoptions seems to be low. The Calico charts and manifest in [projectcalico/calico](https://github.com/projectcalico/calico) run calico-node in privileged mode.

### Seamless overlay network mode switching

**Feature State**: `Alpha`

##### Motivation

When switching Calico from overlay mode (IPIP) to non-overlay mode, there is a critical transition period where pod-to-pod communication can be disrupted if the network routes are not properly configured. In non-overlay mode, Calico relies on the cloud provider's route controller to create routes for pod-to-pod communication. If overlay is disabled before these routes are created, pods may lose connectivity.

##### Support for seamless overlay switching

The `SeamlessOverlaySwitch` feature gate enables validation of node routes before disabling overlay networking. When this feature is enabled and an overlay-to-non-overlay switch is detected, the extension will:

1. Check that all nodes have the `NetworkUnavailable` condition set to `False` with reason `RouteCreated`
2. Only proceed with disabling overlay once routes are confirmed to be in place

This prevents connectivity issues during the transition period. The feature is controlled via feature gate named `SeamlessOverlaySwitch`. The feature gates are configured in the [ControllerConfiguration](../../example/00-componentconfig.yaml) of networking-calico. The corresponding ControllerDeployment configuration that enables the `SeamlessOverlaySwitch` would look like:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: ControllerDeployment
metadata:
  name: networking-calico
type: helm
providerConfig:
  values:
    chart: <omitted>
    config:
      featureGates:
        SeamlessOverlaySwitch: true
```

##### Kubernetes version requirements

The seamless overlay switch relies on the `MutatingAdmissionPolicy` admission API. The availability of this API depends on the shoot's Kubernetes version:

| Kubernetes version | MutatingAdmissionPolicy state | What you need to do |
|--------------------|-------------------------------|---------------------|
| < 1.34             | Alpha (off by default)                                                                                              | Explicitly enable via feature gate and runtimeConfig (see below) |
| >= 1.34, < 1.36    | Beta, but [off by default per KEP-3136](https://github.com/kubernetes/enhancements/tree/master/keps/sig-architecture/3136-beta-apis-off-by-default) | Explicitly enable via feature gate and runtimeConfig (see below) |
| >= 1.36            | GA (always on)                                                                                                      | Nothing — seamless switch activates automatically |

**Enabling MutatingAdmissionPolicy on Kubernetes < 1.36**

For shoots on 1.33 (alpha) or 1.34 / 1.35 (beta, off by default per KEP-3136), the feature must be opted in explicitly. Set the feature gate and the matching `runtimeConfig` entry in the shoot spec:

```yaml
spec:
  kubernetes:
    version: 1.34.3
    kubeAPIServer:
      featureGates:
        MutatingAdmissionPolicy: true
      runtimeConfig:
        admissionregistration.k8s.io/v1alpha1: true
        admissionregistration.k8s.io/v1beta1: true
```

The API is served under `v1alpha1` on 1.33 and promoted to `v1beta1` on 1.34. Enabling both runtimeConfig entries keeps the configuration valid across upgrades between these versions.

**Migrating from Kubernetes 1.35 → 1.36**

On 1.36 the feature graduates to GA and is locked on, so the explicit feature gate and `runtimeConfig` entries are no longer required (and `MutatingAdmissionPolicy: false` is rejected). Remove any explicit overrides before or during the upgrade:

```yaml
spec:
  kubernetes:
    version: 1.36.0
    kubeAPIServer:
      featureGates:
        # Remove or omit any prior MutatingAdmissionPolicy setting
```

##### Behavior

- **`SeamlessOverlaySwitch` enabled**: The extension validates that routes are created before disabling overlay. If routes are not ready, the reconciliation will fail with a retriable error, keeping overlay enabled until routes are confirmed.
- **`SeamlessOverlaySwitch` disabled**: The extension will disable overlay immediately when requested, without checking for route readiness. This may result in temporary connectivity issues during the transition.

##### Limitations

This validation only applies when switching from overlay-enabled to overlay-disabled. It does not affect other configuration changes.

### `kube-apiserver` `GlobalNetworkSet`

The extension can maintain a Calico `GlobalNetworkSet` named `gardener-kube-apiserver` in every shoot cluster, holding the IP address(es) of the shoot's `kube-apiserver` as reachable from within the shoot. Shoot owners reference it from their own Calico policies in order to restrict egress traffic to the `kube-apiserver`, see the [usage documentation](../usage/usage.md#restricting-access-to-the-kube-apiserver).

The feature is disabled by default. It can be enabled landscape-wide in the component configuration:

```yaml
apiVersion: calico.networking.extensions.config.gardener.cloud/v1alpha1
kind: ControllerConfiguration
kubeAPIServerEndpoints:
  enabled: true
```

Shoots override this via `.spec.networking.providerConfig.kubeAPIServerEndpoints.enabled`, so the effective value is `providerConfig.enabled ?? componentConfig.enabled ?? false`.

##### Address source

The addresses are read from the `DNSRecord`s labelled `gardener.cloud/role=controlplane` and `role in (internal, external)` in the shoot's control plane namespace. For `A`/`AAAA` records their `spec.values` already are the IP addresses of the seed's istio ingress gateway load balancer, and they are the write side of the very DNS entry shoot pods resolve - so the published set matches what pods observe, without the extension performing DNS lookups. If no `DNSRecord` exists (unmanaged DNS, local development setups), the kube-apiserver entries of `shoot.status.advertisedAddresses` are used instead.

##### Update timing

The set is recomputed during the `Network` reconciliation, i.e. once per shoot reconciliation (hourly by default, see `controllers.shoot.syncPeriod` in the gardenlet configuration). Nothing watches the `DNSRecord`s.

That is normally sufficient, because `DNSRecord.spec.values` is written by the same flow, which updates it before the `Network`. The exception is a reconciliation failing *after* the `DNSRecord` was updated but *before* the `Network` was reconciled: DNS then points to the new address while the set still holds the previous one, and policy covered pods lose access to the kube-apiserver until the next successful reconciliation. The shoot is in `lastOperation.state: Error` meanwhile. The inverse is harmless - if the `DNSRecord` could not be updated either, DNS and the set stay consistent.

> ⚠️ Should pods be unable to reach the kube-apiserver after a control plane migration, after an `ExposureClass` or high availability change, or after the istio ingress gateway load balancer of a seed was recreated, trigger a reconciliation of the affected shoots: `kubectl -n garden-<project> annotate shoot <name> gardener.cloud/operation=reconcile`. If the shoot's `lastOperation.state` is `Failed`, `gardener.cloud/operation=retry` is required instead - `reconcile` is ignored in that state.

##### Unsupported: `kube-apiserver` exposed via a hostname

Landscapes exposing the istio ingress gateway via a **hostname** instead of an IP address (for example an AWS NLB) cannot be supported: a `GlobalNetworkSet` holds CIDRs, and resolving the hostname is not implemented - the published set would drift from what pods resolve, and keeping it in sync would make the extension a DNS client.

gardenlet derives the `DNSRecord` type from the address it publishes, so such a landscape is recognisable without any lookup: the `DNSRecord`s are of type `CNAME`. In that case the extension **fails the reconciliation** of the `Network` resource with the error code `ERR_CONFIGURATION_PROBLEM`:

```
the kube-apiserver is exposed via a hostname instead of an IP address: [abc.elb.eu-west-1.amazonaws.com],
hence no GlobalNetworkSet can be deployed - unset `kubeAPIServerEndpoints.enabled` for this shoot or
disable it landscape-wide
```

The way out is to disable the feature, either for the shoot or landscape-wide. Note that a landscape-wide default of `enabled: true` therefore fails **every** shoot of such a landscape.

##### If the addresses are not published yet

A missing address is a different, transient situation: during the initial creation the load balancer address is not known yet, and there is nothing to publish. The extension neither fails nor publishes an empty set then, because failing would block the whole shoot flow and an empty set would silently break every policy referencing it. Any previously published set is left untouched - which is also why the set lives in its own `ManagedResource` instead of in the calico chart, from which a missing object would be deleted again.

Since the reconciliation succeeds, the situation is reported as a warning event instead:

```bash
kubectl -n <control-plane-namespace> describe network calico-network
```

| Reason | Meaning |
| --- | --- |
| `KubeAPIServerEndpointsOutdated` | A previously published `GlobalNetworkSet` is still in place. Policies keep working with the addresses of the last successful reconciliation and may be outdated. |
| `KubeAPIServerEndpointsMissing` | No `GlobalNetworkSet` has been published at all, so policies referring to it match no address and **block** traffic to the kube-apiserver. |

##### Inspecting the deployed set

```bash
kubectl -n <control-plane-namespace> get managedresource extension-networking-calico-apiserver-endpoints
```

The rendered object carries the source it was derived from as an annotation, and intentionally no timestamp: its bytes are hashed into an immutable secret, so a timestamp would create a new secret and re-apply the object in every shoot on every reconciliation. Use the secret's `metadata.creationTimestamp` to see when the addresses last changed.

On a newly created shoot this `ManagedResource` can briefly report `ResourcesApplied=False` with `no matches for kind "GlobalNetworkSet"`, and with it `SystemComponentsHealthy=False` on the shoot, until the CRD is established - it ships in the calico chart, i.e. in a different `ManagedResource` which the gardener-resource-manager applies independently. No action is required, it retries.
