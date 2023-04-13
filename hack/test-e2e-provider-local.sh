#!/bin/bash

# SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -o errexit

repo_root="$(readlink -f $(dirname ${0})/..)"

if [[ ! -d "$repo_root/gardener" ]]; then
  git clone https://github.com/gardener/gardener.git
fi

cd "$repo_root/gardener"
git checkout 1f94c1e5b4b9e7cdf40d2e314dc74fea54a0d293 # g/g v1.67.1
source "$repo_root/gardener/hack/ci-common.sh"
make kind-up
trap '{
  cd "$repo_root/gardener"
  export_artifacts "gardener-local"
  make kind-down
}' EXIT
export KUBECONFIG=$repo_root/gardener/example/gardener-local/kind/local/kubeconfig
make gardener-up

cd $repo_root

version=$(git rev-parse HEAD)
make docker-images
docker tag eu.gcr.io/gardener-project/gardener/extensions/networking-calico:latest networking-calico-local:$version
kind load docker-image networking-calico-local:$version --name gardener-local

mkdir -p $repo_root/tmp
cp -f $repo_root/example/controller-registration.yaml $repo_root/tmp/controller-registration.yaml
yq -i e "(select (.providerConfig.values.image) | .providerConfig.values.image.tag) |= \"$version\"" $repo_root/tmp/controller-registration.yaml
yq -i e '(select (.providerConfig.values.image) | .providerConfig.values.image.repository) |= "docker.io/library/networking-calico-local"' $repo_root/tmp/controller-registration.yaml

kubectl apply -f "$repo_root/tmp/controller-registration.yaml"

echo '127.0.0.1 api.e2e-default.local.external.local.gardener.cloud' >> /etc/hosts

# reduce flakiness in contended pipelines
export GOMEGA_DEFAULT_EVENTUALLY_TIMEOUT=5s
export GOMEGA_DEFAULT_EVENTUALLY_POLLING_INTERVAL=200ms
# if we're running low on resources, it might take longer for tested code to do something "wrong"
# poll for 5s to make sure, we're not missing any wrong action
export GOMEGA_DEFAULT_CONSISTENTLY_DURATION=5s
export GOMEGA_DEFAULT_CONSISTENTLY_POLLING_INTERVAL=200ms

GO111MODULE=on ginkgo --timeout=1h --v --progress "$@" $repo_root/test/e2e/...

cd "$repo_root/gardener"
make gardener-down
