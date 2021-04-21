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
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/config"
	"github.com/SENERGY-Platform/process-deployment/lib/ctrl"
	"github.com/SENERGY-Platform/process-deployment/lib/ctrl/deployment/parser"
	"github.com/SENERGY-Platform/process-deployment/lib/ctrl/deployment/stringifier"
	"github.com/SENERGY-Platform/process-deployment/lib/devices"
	"github.com/SENERGY-Platform/process-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-deployment/lib/model/processmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/processrepo"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
)

func init() {
	config.NewId = func() string {
		return "unused-id"
	}
}

type Controller struct {
	config                configuration.Config
	reusedConfig          config.Config
	processrepo           interfaces.ProcessRepo
	deploymentParser      interfaces.DeploymentParser
	deploymentStringifier interfaces.DeploymentStringifier
	deviceRepoFactory     DeviceRepoFactory
	processSync           ProcessSync
	reusedDeviceRepo      interfaces.Devices
}

type ProcessSync interface {
	Deploy(token string, hubId string, deployment deploymentmodel.Deployment) error
}

type ProcessRepo interface {
	GetProcessModel(token string, id string) (result processmodel.ProcessModel, err error, errCode int)
}

type DeviceRepoFactory func(config configuration.Config, reuse interfaces.Devices, hubId string) interfaces.Devices

func New(conf configuration.Config, processSync ProcessSync, deviceRepoFactory DeviceRepoFactory) (*Controller, error) {
	reusedConfig := &config.ConfigStruct{
		ApiPort:                     conf.ApiPort,
		DeviceRepoUrl:               conf.DeviceRepoUrl,
		ProcessRepoUrl:              conf.ProcessRepoUrl,
		PermSearchUrl:               conf.PermSearchUrl,
		DeviceSelectionUrl:          conf.DeviceSelectionUrl,
		Debug:                       conf.Debug,
		NotificationUrl:             conf.NotificationUrl,
		EnableDeviceGroupsForTasks:  conf.EnableDeviceGroupsForTasks,
		EnableDeviceGroupsForEvents: conf.EnableDeviceGroupsForEvents,
		DeploymentTopic:             "deployment-topic-replacement",
	}

	processrepo, err := processrepo.Factory.New(context.Background(), reusedConfig)
	if err != nil {
		return nil, err
	}

	reusedDeviceRepo, err := devices.Factory.New(context.Background(), reusedConfig)
	if err != nil {
		return nil, err
	}
	return &Controller{
		config:                conf,
		reusedConfig:          reusedConfig,
		processrepo:           processrepo,
		deploymentParser:      parser.New(reusedConfig),
		deploymentStringifier: stringifier.New(reusedConfig),
		deviceRepoFactory:     deviceRepoFactory,
		processSync:           processSync,
		reusedDeviceRepo:      reusedDeviceRepo,
	}, nil
}

func (this *Controller) ReuseCloudDeploymentWithProcessSync(token string, hubId string) *ctrl.Ctrl {
	result, _ := ctrl.New(
		context.Background(),
		this.reusedConfig,
		&SourcingReplacement{
			token:       token,
			hubId:       hubId,
			processSync: this.processSync,
		},
		nil,
		this.deviceRepoFactory(this.config, this.reusedDeviceRepo, hubId),
		nil,
		ImportsMock{})
	return result
}

func (this *Controller) ReuseCloudDeploymentWithNewDeviceRepo(hubId string) *ctrl.Ctrl {
	result, _ := ctrl.New(context.Background(), this.reusedConfig, &SourcingReplacement{}, nil, this.deviceRepoFactory(this.config, this.reusedDeviceRepo, hubId), nil, ImportsMock{})
	return result
}

func (this *Controller) ReuseCloudDeployment() *ctrl.Ctrl {
	result, _ := ctrl.New(context.Background(), this.reusedConfig, &SourcingReplacement{}, nil, nil, nil, ImportsMock{})
	return result
}

type ImportsMock struct{}

func (this ImportsMock) CheckAccess(token jwt_http_router.JwtImpersonate, ids []string, alsoCheckTypes bool) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	return false, errors.New("imports not supported for fog processes")
}
