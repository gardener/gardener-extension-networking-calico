#!/usr/bin/env bash
#
# diff-calico-upstream.sh — compare upstream Calico's self-hosted install
# manifest to guide porting RBAC and workload changes into this extension's
# chart on a version bump.
#
# Source manifest: manifests/calico-typha.yaml — the "Kubernetes API datastore,
# more than 50 nodes" (Typha-enabled) variant. This is the canonical reference
# per the internal guide (github.tools.sap/kubernetes/networking-backlog,
# docs/update-calico.md); it carries the calico-typha Deployment and the
# calico-tier-getter ClusterRole that the plain manifests/calico.yaml omits.
#
# Modes:
#
#   (default)  UPSTREAM-OLD vs UPSTREAM-NEW — isolates "what Calico changed"
#              between two releases (RBAC rules, pod specs, calico-config).
#
#   --vscode   Open each relevant chart file in the VS Code diff editor for
#   (--code)   hand-merging: LEFT = the full upstream object (raw), RIGHT = the
#              real chart file (editable — saving merges into the repo). The 4
#              ClusterRoles are opened only when their .rules actually differ; the
#              3 workloads (calico-node DaemonSet, kube-controllers + typha
#              Deployments) and the calico-config ConfigMap are always opened as
#              raw full-object diffs — they are heavily Helm-templated, so there
#              is no clean semantic trigger: expect templating noise and apply
#              judgment. CRDs are excluded (regenerate with update-calico-crds.sh),
#              as are the Gardener-only autoscalers/VPAs.
#
# Map every delta with the carry/drop filter (see calico-bump-mapping.md in the
# workspace root). CRDs are NOT covered here — regenerate them with
# update-calico-crds.sh.
#
# Usage:
#   hack/diff-calico-upstream.sh [OLD_VERSION] [NEW_VERSION]
#   hack/diff-calico-upstream.sh --vscode [NEW_VERSION]
#
# Version defaults (read the calico-node image tag):
#   NEW_VERSION = tag in imagevector/images.yaml (working tree)
#   OLD_VERSION = tag in imagevector/images.yaml at git HEAD
#
# Requires: curl, yq (mikefarah v4), git. (--vscode also needs the `code` CLI.)

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
images_file="${repo_root}/imagevector/images.yaml"

for bin in curl yq git; do
  command -v "$bin" >/dev/null || { echo "error: '$bin' is required" >&2; exit 1; }
done
yq --version 2>&1 | grep -qE 'version v?4' || { echo "error: yq v4 (mikefarah) is required" >&2; exit 1; }

mode="upstream"
case "${1:-}" in
  --vscode|--code)   mode="vscode"; shift ;;
  --*)               echo "error: unknown option '$1' (use --vscode)" >&2; exit 1 ;;
esac

node_tag_expr='.images[] | select(.name == "calico-node") | .tag'
default_new="$(yq "$node_tag_expr" "$images_file")"

if [[ "$mode" == "upstream" ]]; then
  old_version="${1:-$(git -C "$repo_root" show HEAD:imagevector/images.yaml | yq "$node_tag_expr" -)}"
  new_version="${2:-$default_new}"
  [[ -n "$old_version" ]] || { echo "error: could not determine OLD_VERSION" >&2; exit 1; }
else
  old_version=""
  new_version="${1:-$default_new}"
fi
[[ -n "$new_version" ]] || { echo "error: could not determine NEW_VERSION" >&2; exit 1; }

source_repo="$(yq '.images[] | select(.name == "calico-node") | .sourceRepository' "$images_file")"
# Normalize to "owner/repo": strip an optional scheme and the github.com host, plus any
# trailing slash. sourceRepository is "github.com/projectcalico/calico" (no scheme), but
# tolerate an "https://github.com/..." form too.
gh_path="${source_repo#https://}"
gh_path="${gh_path#github.com/}"
gh_path="${gh_path%/}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

fetch() { # version -> file
  local version="$1" out="$2"
  local url="https://raw.githubusercontent.com/${gh_path}/${version}/manifests/calico-typha.yaml"
  curl -fsSL "$url" -o "$out" || { echo "error: failed to fetch $url" >&2; exit 1; }
}

new_file="$tmp/calico-new.yaml"
fetch "$new_version" "$new_file"

# Helper: open a VS Code diff or print the command if `code` CLI is unavailable.
open_diff() {
  local up_obj="$1" chart_file="$2"
  if command -v code >/dev/null; then
    code --diff "$up_obj" "$chart_file"
  else
    echo "      code --diff \"$up_obj\" \"$chart_file\""
  fi
}

# ---- vscode mode -----------------------------------------------------------
if [[ "$mode" == "vscode" ]]; then
  # LEFT pane  = the full upstream object (raw), written to a persistent dir (NOT
  #              under $tmp) so the diff tab stays valid after this script exits.
  # RIGHT pane = the real chart file (editable; saving merges into the repo).
  outdir="$(mktemp -d "${TMPDIR:-/tmp}/calico-rbac-diff.XXXXXX")"
  echo "VS Code merge:  LEFT = upstream ${new_version} (raw)   RIGHT = real chart file (edit here)"
  echo "Upstream reference objects: $outdir"
  command -v code >/dev/null 2>&1 || echo "('code' CLI not found — run the printed commands manually.)"
  echo

  # ClusterRoles: opened only when .rules differ (semantic trigger).
  echo "# ClusterRoles (opened only when their .rules differ):"
  for role in calico-node calico-cni-plugin calico-kube-controllers calico-tier-getter; do
    case "$role" in
      calico-node)             chart_file="$repo_root/charts/internal/calico/templates/node/clusterrole-calico-node.yaml" ;;
      calico-cni-plugin)       chart_file="$repo_root/charts/internal/calico/templates/node/clusterrole-calico-cni-plugin.yaml" ;;
      calico-kube-controllers) chart_file="$repo_root/charts/internal/calico/templates/kube-controller/clusterrole-calico-kube-controllers.yaml" ;;
      calico-tier-getter)      chart_file="$repo_root/charts/internal/calico/templates/tier-getter/clusterrole-tier-getter.yaml" ;;
    esac
    # Strip Helm template lines so the file is valid YAML for yq parsing.
    grep -v '{{' "$chart_file" | yq '.rules' - > "$tmp/ext.rules" 2>/dev/null || true
    yq eval-all "select(.kind == \"ClusterRole\" and .metadata.name == \"$role\") | .rules" \
      "$new_file" > "$tmp/up.rules" 2>/dev/null || true
    if diff -q "$tmp/ext.rules" "$tmp/up.rules" >/dev/null 2>&1; then
      printf "  %-26s identical rules (not opened)\n" "$role"
      continue
    fi
    up_obj="$outdir/${role}.upstream-${new_version}.yaml"
    yq eval-all "select(.kind == \"ClusterRole\" and .metadata.name == \"$role\")" \
      "$new_file" > "$up_obj" 2>/dev/null || true
    printf "  %-26s DIFFERS -> opening\n" "$role"
    open_diff "$up_obj" "$chart_file"
  done

  # Workloads + config: always opened (raw full-object diffs — expect templating noise).
  echo
  echo "# Workloads + Felix/CNI config (raw full-object diffs — expect templating noise):"
  declare -a wk_ents=(
    "DaemonSet|calico-node|charts/internal/calico/templates/node/daemonset-calico-node.yaml"
    "Deployment|calico-kube-controllers|charts/internal/calico/templates/kube-controller/deployment-calico-kube-controllers.yaml"
    "Deployment|calico-typha|charts/internal/calico/templates/typha/deployment-calico-typha.yaml"
    "ConfigMap|calico-config|charts/internal/calico/templates/node/configmap-calico-config.yaml"
  )
  for entry in "${wk_ents[@]}"; do
    IFS='|' read -r kind name chart <<< "$entry"
    chart_file="$repo_root/$chart"
    if [[ ! -f "$chart_file" ]]; then
      printf "  %-40s chart file missing (skipped): %s\n" "$kind/$name" "$chart"; continue
    fi
    up_obj="$outdir/${kind}-${name}.upstream-${new_version}.yaml"
    yq eval-all "select(.kind == \"$kind\" and .metadata.name == \"$name\")" \
      "$new_file" > "$up_obj" 2>/dev/null || true
    if [[ ! -s "$up_obj" ]]; then
      printf "  %-40s not in upstream manifest (skipped)\n" "$kind/$name"; continue
    fi
    printf "  %-40s opening\n" "$kind/$name"
    open_diff "$up_obj" "$chart_file"
  done

  echo
  echo "Edit the RIGHT pane (your repo file) to merge; saving writes to the repo."
  echo "Upstream refs kept at $outdir (delete when done)."
  exit 0
fi

# ---- default mode: upstream-old vs upstream-new ----------------------------
old_file="$tmp/calico-old.yaml"
fetch "$old_version" "$old_file"

echo "Diffing upstream ${gh_path} manifests/calico-typha.yaml:  ${old_version}  ->  ${new_version}"
echo "Reminder: this is raw upstream. Apply the carry/drop filter before porting."

diff_section() { # title  yq-expression
  local title="$1" expr="$2"
  yq eval-all "$expr" "$old_file" > "$tmp/old.txt" 2>/dev/null || true
  yq eval-all "$expr" "$new_file" > "$tmp/new.txt" 2>/dev/null || true
  echo
  echo "===================================================================="
  echo "  $title"
  echo "===================================================================="
  if [[ ! -s "$tmp/old.txt" && ! -s "$tmp/new.txt" ]]; then
    echo "  (not present in these manifests)"
  elif diff -q "$tmp/old.txt" "$tmp/new.txt" >/dev/null 2>&1; then
    echo "  (no change)"
  else
    git --no-pager diff --no-index -- "$tmp/old.txt" "$tmp/new.txt" || true
  fi
}

# --- RBAC -------------------------------------------------------------------
diff_section "ClusterRole: calico-node  (.rules)" \
  'select(.kind == "ClusterRole" and .metadata.name == "calico-node") | .rules'
diff_section "ClusterRole: calico-cni-plugin  (.rules)" \
  'select(.kind == "ClusterRole" and .metadata.name == "calico-cni-plugin") | .rules'
diff_section "ClusterRole: calico-kube-controllers  (.rules)" \
  'select(.kind == "ClusterRole" and .metadata.name == "calico-kube-controllers") | .rules'
diff_section "ClusterRole: calico-tier-getter  (.rules)" \
  'select(.kind == "ClusterRole" and .metadata.name == "calico-tier-getter") | .rules'

# --- Workload manifests -----------------------------------------------------
diff_section "DaemonSet: calico-node  (.spec.template.spec)" \
  'select(.kind == "DaemonSet" and .metadata.name == "calico-node") | .spec.template.spec'
diff_section "Deployment: calico-kube-controllers  (.spec.template.spec)" \
  'select(.kind == "Deployment" and .metadata.name == "calico-kube-controllers") | .spec.template.spec'
diff_section "Deployment: calico-typha  (.spec.template.spec)" \
  'select(.kind == "Deployment" and .metadata.name == "calico-typha") | .spec.template.spec'

# --- Felix / CNI configuration ----------------------------------------------
diff_section "ConfigMap: calico-config  (.data)" \
  'select(.kind == "ConfigMap" and .metadata.name == "calico-config") | .data'

echo
echo "Done. Note: upstream image tags inside pod specs will always differ (${old_version} -> ${new_version}) — that is expected noise."
