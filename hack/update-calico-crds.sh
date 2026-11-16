#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0
#
# Re-syncs the vendored Calico CRDs under
#   charts/internal/calico/templates/custom-resource-definition/
# from the upstream projectcalico/calico release manifest, applying the
# Gardener-specific transformations:
#   - strip the upstream banner and per-document "# Source:" comments
#   - strip the controller-gen version annotation; keep all other upstream annotations,
#     in particular api-approved.kubernetes.io (required for CRDs in *.k8s.io groups)
#   - add the "gardener.cloud/role: system-component" label
#
# Usage:
#   hack/update-calico-crds.sh [VERSION]
#
# VERSION is the upstream release tag (with or without a leading "v", e.g.
# "v3.32.2"). If omitted, it defaults to the calico-node image tag in
# imagevector/images.yaml, so a plain invocation after the image bump re-syncs
# the matching CRDs.
#
# NOTE: The set of CRDs is version-dependent. Up to v3.31 the manifest includes the
# AdminNetworkPolicy CRDs (adminnetworkpolicies, baselineadminnetworkpolicies, group
# policy.networking.k8s.io); v3.32 removes those and adds clusternetworkpolicies
# (same group). This script writes exactly the CRDs present in the target release's
# manifest and never deletes files, so a minor bump may leave now-obsolete CRD files
# behind for manual removal (see the bump plan).

set -o errexit
set -o nounset
set -o pipefail

repo_root="$(git rev-parse --show-toplevel)"
crd_dir="${repo_root}/charts/internal/calico/templates/custom-resource-definition"
images_file="${repo_root}/imagevector/images.yaml"

for tool in yq curl; do
  if ! command -v "${tool}" >/dev/null 2>&1; then
    echo "error: required tool '${tool}' not found in PATH" >&2
    exit 1
  fi
done

version="${1:-$(yq eval '.images[] | select(.name == "calico-node") | .tag' "${images_file}")}"
# Normalise to a leading-v git tag (quay publishes both "3.32.2" and "v3.32.2").
ref="v${version#v}"

# Derive the upstream repository from images.yaml so there is a single source of truth.
source_repository="$(yq eval '.images[] | select(.name == "calico-node") | .sourceRepository' "${images_file}")"
# e.g. github.com/projectcalico/calico -> projectcalico/calico
owner_repo="${source_repository#github.com/}"

manifest_url="https://raw.githubusercontent.com/${owner_repo}/${ref}/manifests/crds.yaml"

echo "Re-syncing Calico CRDs for ${ref}"
echo "  source: ${manifest_url}"
echo "  target: ${crd_dir}"

manifest="$(mktemp)"
trap 'rm -f "${manifest}"' EXIT
# Fetch and drop the upstream banner + per-document "# Source:" comments. These are
# stripped textually (rather than via yq) so that yq is not forced to recurse over
# every node, which would reset scalar styles and reflow multi-line descriptions.
# Inline schema comments are deliberately preserved.
curl --fail --silent --show-error --location "${manifest_url}" \
  | grep -vE '^# (Source:|CustomResourceDefinitions for Calico)' \
  > "${manifest}"

# Map an upstream plural to the chart's file name. All CRDs follow the
# crd-<plural>.yaml convention except kubecontrollersconfigurations, which this
# chart has historically stored under its singular name.
crd_filename() {
  case "$1" in
    kubecontrollersconfigurations) echo "crd-kubecontrollersconfiguration.yaml" ;;
    *)                             echo "crd-$1.yaml" ;;
  esac
}

count=0
# Newline-separated list of the CRD files this release expects, used below to
# detect obsolete files left over from a previous (larger) CRD set.
expected_files=""
while IFS= read -r plural; do
  [[ -n "${plural}" ]] || continue
  filename="$(crd_filename "${plural}")"
  expected_files="${expected_files}${filename}
"
  {
    echo "---"
    # Transform each CRD:
    #   - strip only the controller-gen version annotation; keep every other upstream
    #     annotation, in particular api-approved.kubernetes.io (required by Kubernetes
    #     for CRDs in *.k8s.io groups such as policy.networking.k8s.io) and the
    #     network-policy-api bundle-version/channel provenance. Drop the annotations
    #     map if it became empty.
    #   - add the Gardener system-component label
    # yq emits a leading "---" for the first document in the stream; the sed drops it so
    # the echo above is the sole separator.
    yq eval "
      select(.spec.names.plural == \"${plural}\")
      | del(.metadata.annotations.\"controller-gen.kubebuilder.io/version\")
      | del(.metadata.annotations | select(length == 0))
      | .metadata.labels.\"gardener.cloud/role\" = \"system-component\"
    " "${manifest}" | sed '1{/^---$/d;}'
  } > "${crd_dir}/${filename}"
  echo "  wrote ${filename}"
  count=$((count + 1))
done < <(yq eval -N '.spec.names.plural' "${manifest}")

echo "Done: re-synced ${count} CRD(s)."

# Detect chart CRD files that no longer correspond to any CRD in the target
# release (removed or renamed upstream, e.g. the AdminNetworkPolicy CRDs dropped
# in v3.32). This script never deletes files, and an obsolete file is otherwise
# invisible: it is left unchanged, so it does not show up in the diff and keeps
# being deployed. Fail loudly instead, so the removal is handled deliberately
# (deleting a CRD drops all of its custom resources on the next deployment and
# must be a conscious, reviewed action).
orphans=()
for existing in "${crd_dir}"/crd-*.yaml; do
  [[ -e "${existing}" ]] || continue
  base="$(basename "${existing}")"
  if ! grep -qxF "${base}" <<<"${expected_files}"; then
    orphans+=("${base}")
  fi
done

if [[ ${#orphans[@]} -gt 0 ]]; then
  echo
  echo "error: ${#orphans[@]} vendored CRD file(s) are not present in ${ref} and are now obsolete:" >&2
  printf '  - %s\n' "${orphans[@]}" >&2
  echo >&2
  echo "They were left untouched (this script never deletes). After confirming nothing" >&2
  echo "relies on the removed CRD(s), delete them with 'git rm' and re-run this script." >&2
  exit 1
fi

echo
echo "Next steps:"
echo "  - review the diff (new CRDs, schema and RBAC-relevant resource changes)"
echo "  - the AdminNetworkPolicy CRDs (group policy.networking.k8s.io) are managed separately"
