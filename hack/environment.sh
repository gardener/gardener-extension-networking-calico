#!/bin/bash
#
# Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

get_cd_registry () {
  echo "eu.gcr.io/sap-se-gcr-k8s-private/cnudie/gardener/development"
}

get_cd_component_name () {
  echo "github.com/gardener/gardener-extension-networking-calico"
}

get_image_registry () {
  echo "eu.gcr.io/gardener-project/gardener/extensions"
}