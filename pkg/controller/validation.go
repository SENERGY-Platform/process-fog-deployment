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

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
)

func validateDeployment(msg messages.DeploymentCommand) error {
	if msg.Deployment == nil {
		return errors.New("expect deployment in ProducerReplacement.Produce()")
	}
	if msg.Version != deploymentmodel.CurrentVersion {
		return errors.New("unexpected deployment version")
	}
	for _, element := range msg.Deployment.Elements {
		if element.MessageEvent != nil {
			return errors.New("fog process deployments dont support message events. please use conditional events")
		}
		if element.ConditionalEvent != nil && element.ConditionalEvent.Selection.SelectedImportId != nil {
			return errors.New("fog process deployments dont support imports as selection for conditional events")
		}
	}
	return nil
}
