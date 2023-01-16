/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"time"
)

type DeploymentMetadata struct {
	Metadata
	SyncInfo
}

type Metadata struct {
	CamundaDeploymentId string                     `json:"camunda_deployment_id"`
	ProcessParameter    map[string]Variable        `json:"process_parameter"`
	DeploymentModel     deploymentmodel.Deployment `json:"deployment_model"`
}

type SyncInfo struct {
	NetworkId       string    `json:"network_id"`
	IsPlaceholder   bool      `json:"is_placeholder"`
	MarkedForDelete bool      `json:"marked_for_delete"`
	SyncDate        time.Time `json:"sync_date"`
}

type Variable struct {
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`
	ValueInfo interface{} `json:"valueInfo"`
}
