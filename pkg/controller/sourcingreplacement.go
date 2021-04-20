/*
 * Copyright 2021 InfAI (CC SES)
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
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/config"
	"github.com/SENERGY-Platform/process-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
)

//mocks sourcing interface to reuse github.com/SENERGY-Platform/process-deployment/lib/ctrl without connecting to kafka
type SourcingReplacement struct {
	token       string
	hubId       string
	processSync ProcessSync
	devicerepo  interfaces.Devices
}

func (this *SourcingReplacement) NewConsumer(ctx context.Context, config config.Config, topic string, listener func(delivery []byte) error) error {
	return nil
}

// reroutes deployment requests to github.com/SENERGY-Platform/process-sync
type ProducerReplacement struct {
	token       string
	hubId       string
	processSync ProcessSync
	devicerepo  interfaces.Devices
}

func (this *ProducerReplacement) Produce(topic string, message []byte) error {
	deplMsg := messages.DeploymentCommand{}
	err := json.Unmarshal(message, &deplMsg)
	if err != nil {
		return err
	}
	if deplMsg.DeploymentV2 == nil {
		return errors.New("expect deployment v2 in ProducerReplacement.Produce()")
	}
	fogDeplMessage, err := this.createFogDeploymentMessage(this.token, *deplMsg.DeploymentV2)
	if err != nil {
		return err
	}
	return this.processSync.Deploy(this.token, this.hubId, fogDeplMessage)
}

func (this *ProducerReplacement) createFogDeploymentMessage(token string, deployment deploymentmodel.Deployment) (result model.FogDeploymentMessage, err error) {
	result.Deployment = deployment
	result.Devices = map[string]devicemodel.Device{}
	result.Services = map[string]devicemodel.Service{}
	for _, element := range deployment.Elements {
		var selection deploymentmodel.Selection
		if element.MessageEvent != nil {
			selection = element.MessageEvent.Selection
		}
		if element.Task != nil {
			selection = element.Task.Selection
		}
		if selection.SelectedDeviceId == nil {
			return result, errors.New("expect selected device id")
		}
		if selection.SelectedServiceId == nil {
			return result, errors.New("expect selected service id")
		}
		device, err, _ := this.devicerepo.GetDevice(jwt_http_router.JwtImpersonate(token), *selection.SelectedDeviceId)
		if err != nil {
			return result, err
		}
		service, err, _ := this.devicerepo.GetService(jwt_http_router.JwtImpersonate(token), *selection.SelectedServiceId)
		if err != nil {
			return result, err
		}
		result.Devices[device.Id] = device
		result.Services[service.Id] = service
	}
	return result, nil
}

func (this *SourcingReplacement) NewProducer(ctx context.Context, config config.Config, topic string) (interfaces.Producer, error) {
	return &ProducerReplacement{
		token:       this.token,
		hubId:       this.hubId,
		processSync: this.processSync,
		devicerepo:  this.devicerepo,
	}, nil
}
