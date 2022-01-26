// Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"github.com/gardener/gardener-extension-networking-calico/pkg/calico"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var logger = log.Log.WithName("calico-validator-webhook")

// New creates a new webhook that validates Shoot resources.
func New(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	logger.Info("Setting up webhook", "name", extensionswebhook.ValidatorName)

	return extensionswebhook.New(mgr, extensionswebhook.Args{
		Provider:   calico.Name,
		Name:       extensionswebhook.ValidatorName,
		Path:       extensionswebhook.ValidatorPath,
		Predicates: []predicate.Predicate{createCalicoPredicate()},
		Validators: map[extensionswebhook.Validator][]extensionswebhook.Type{
			NewShootValidator(): {{Obj: &core.Shoot{}}},
		},
	})
}

func createCalicoPredicate() predicate.Funcs {
	f := func(obj client.Object) bool {
		if obj == nil {
			return false
		}

		shoot, ok := obj.(*core.Shoot)
		if !ok {
			return false
		}

		return shoot.Spec.Networking.Type == calico.ReleaseName
	}

	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return f(event.Object)
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return f(event.ObjectNew)
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return f(event.Object)
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return f(event.Object)
		},
	}
}
