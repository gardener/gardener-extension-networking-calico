// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "$TOOLS_BIN_DIR/extension-generator --name=networking-calico --provider-type=calico --component-category=network --extension-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/networking-calico:$(cat ../VERSION) --admission-runtime-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/admission-calico-runtime:$(cat ../VERSION) --admission-application-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/admission-calico-application:$(cat ../VERSION) --destination=./extension/base/extension.yaml"
//go:generate sh -c "$TOOLS_BIN_DIR/kustomize build ./extension -o ./extension.yaml"

package example
