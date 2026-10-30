#!/usr/bin/env bash
# Reads CNI_PLUGINS_VERSION from the Dockerfile and updates the per-architecture
# SHA256 ARG defaults to match the upstream release checksums.
# Run this whenever CNI_PLUGINS_VERSION is bumped.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKERFILE="${SCRIPT_DIR}/../Dockerfile"

VERSION="$(grep -E '^ARG CNI_PLUGINS_VERSION=' "${DOCKERFILE}" | cut -d= -f2)"
if [[ -z "${VERSION}" ]]; then
  echo "ERROR: could not determine CNI_PLUGINS_VERSION from ${DOCKERFILE}" >&2
  exit 1
fi
echo "Updating CNI plugins checksums for ${VERSION}..."

fetch_sha256() {
  local arch="${1}"
  local url="https://github.com/containernetworking/plugins/releases/download/${VERSION}/cni-plugins-linux-${arch}-${VERSION}.tgz.sha256"
  curl -fsSL "${url}" | awk '{print $1}'
}

AMD64_SHA="$(fetch_sha256 amd64)"
ARM64_SHA="$(fetch_sha256 arm64)"

tmpfile="$(mktemp)"
sed "s|^ARG CNI_PLUGINS_SHA256_AMD64=.*|ARG CNI_PLUGINS_SHA256_AMD64=${AMD64_SHA}|" "${DOCKERFILE}" |
  sed "s|^ARG CNI_PLUGINS_SHA256_ARM64=.*|ARG CNI_PLUGINS_SHA256_ARM64=${ARM64_SHA}|" > "${tmpfile}"
mv "${tmpfile}" "${DOCKERFILE}"

echo "  amd64: ${AMD64_SHA}"
echo "  arm64: ${ARM64_SHA}"
