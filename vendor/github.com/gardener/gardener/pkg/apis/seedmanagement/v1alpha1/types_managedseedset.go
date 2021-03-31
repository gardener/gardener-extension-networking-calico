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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedSeedSet represents a set of identical ManagedSeeds.
type ManagedSeedSet struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec defines the desired identities of ManagedSeeds and Shoots in this set.
	// +optional
	Spec ManagedSeedSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Status is the current status of ManagedSeeds and Shoots in this ManagedSeedSet.
	// +optional
	Status ManagedSeedSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedSeedSetList is a list of ManagedSeed objects.
type ManagedSeedSetList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list object metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is the list of ManagedSeedSets.
	Items []ManagedSeedSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// ManagedSeedSetSpec is the specification of a ManagedSeedSet.
type ManagedSeedSetSpec struct {
	// Replicas is the desired number of replicas of the given Template. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
	// Selector is a label query over ManagedSeeds and Shoots that should match the replica count.
	// It must match the ManagedSeeds and Shoots template's labels.
	Selector metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`
	// Template describes the ManagedSeed that will be created if insufficient replicas are detected.
	// Each ManagedSeed created / updated by the ManagedSeedSet will fulfill this template.
	Template ManagedSeedTemplate `json:"template" protobuf:"bytes,3,opt,name=template"`
	// ShootTemplate describes the Shoot that will be created if insufficient replicas are detected for hosting the corresponding ManagedSeed.
	// Each Shoot created / updated by the ManagedSeedSet will fulfill this template.
	ShootTemplate gardencorev1beta1.ShootTemplate `json:"shootTemplate" protobuf:"bytes,4,rep,name=shootTemplate"`
	// UpdateStrategy specifies the UpdateStrategy that will be
	// employed to update ManagedSeeds / Shoots in the ManagedSeedSet when a revision is made to
	// Template / ShootTemplate.
	// +optional
	UpdateStrategy *UpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,5,opt,name=updateStrategy"`
	// RevisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the ManagedSeedSet's revision history. Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`
}

// UpdateStrategy specifies the strategy that the ManagedSeedSet
// controller will use to perform updates. It includes any additional parameters
// necessary to perform the update for the indicated strategy.
type UpdateStrategy struct {
	// Type indicates the type of the UpdateStrategy. Defaults to RollingUpdate.
	// +optional
	Type *UpdateStrategyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=UpdateStrategyType"`
	// RollingUpdate is used to communicate parameters when Type is RollingUpdateStrategyType.
	// +optional
	RollingUpdate *RollingUpdateStrategy `json:"rollingUpdate,omitempty" protobuf:"bytes,2,opt,name=rollingUpdate"`
}

// UpdateStrategyType is a string enumeration type that enumerates
// all possible update strategies for the ManagedSeedSet controller.
type UpdateStrategyType string

const (
	// RollingUpdateStrategyType indicates that update will be
	// applied to all ManagedSeeds / Shoots in the ManagedSeedSet with respect to the ManagedSeedSet
	// ordering constraints.
	RollingUpdateStrategyType UpdateStrategyType = "RollingUpdate"
)

// RollingUpdateStrategy is used to communicate parameters for RollingUpdateStrategyType.
type RollingUpdateStrategy struct {
	// Partition indicates the ordinal at which the ManagedSeedSet should be partitioned. Defaults to 0.
	// +optional
	Partition *int32 `json:"partition,omitempty" protobuf:"varint,1,opt,name=partition"`
}

// ManagedSeedSetStatus represents the current state of a ManagedSeedSet.
type ManagedSeedSetStatus struct {
	// ObservedGeneration is the most recent generation observed for this ManagedSeedSet. It corresponds to the
	// ManagedSeedSet's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// Replicas is the number of replicas (ManagedSeeds and their corresponding Shoots) created by the ManagedSeedSet controller.
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`
	// ReadyReplicas is the number of ManagedSeeds created by the ManagedSeedSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`
	// NextReplicaNumber is the ordinal number that will be assigned to the next replica of the ManagedSeedSet.
	NextReplicaNumber int32 `json:"nextReplicaNumber,omitempty" protobuf:"bytes,4,opt,name=nextReplicaNumber"`
	// CurrentReplicas is the number of ManagedSeeds created by the ManagedSeedSet controller from the ManagedSeedSet version
	// indicated by CurrentRevision.
	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,5,opt,name=currentReplicas"`
	// UpdatedReplicas is the number of ManagedSeeds created by the ManagedSeedSet controller from the ManagedSeedSet version
	// indicated by UpdateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,6,opt,name=updatedReplicas"`
	// CurrentRevision, if not empty, indicates the version of the ManagedSeedSet used to generate ManagedSeeds with smaller
	// ordinal numbers during updates.
	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,7,opt,name=currentRevision"`
	// UpdateRevision, if not empty, indicates the version of the ManagedSeedSet used to generate ManagedSeeds with larger
	// ordinal numbers during updates
	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,8,opt,name=updateRevision"`
	// CollisionCount is the count of hash collisions for the ManagedSeedSet. The ManagedSeedSet controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,9,opt,name=collisionCount"`
	// Conditions represents the latest available observations of a ManagedSeedSet's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []gardencorev1beta1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`
}

// TODO Condition constants
