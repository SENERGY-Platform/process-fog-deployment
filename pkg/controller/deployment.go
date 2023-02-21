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
	"github.com/SENERGY-Platform/process-deployment/lib/auth"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"net/http"
	"net/url"
)

func (this *Controller) PrepareDeployment(token string, hubId string, xml string, svg string) (result deploymentmodel.Deployment, err error, code int) {
	result, err = this.deploymentParser.PrepareDeployment(xml)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	err = this.ReuseCloudDeploymentWithNewDeviceRepo(hubId).SetDeploymentOptions(auth.Token{Token: token}, &result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result.Diagram.Svg = svg
	this.SetExecutableFlag(&result)
	result.IncidentHandling = &deploymentmodel.IncidentHandling{
		Restart: false,
		Notify:  true,
	}
	return result, nil, http.StatusOK
}

func (this *Controller) CreateDeployment(token string, hubId string, deployment deploymentmodel.Deployment, source string, optionals map[string]bool) (result deploymentmodel.Deployment, err error, code int) {
	jwtToken, err := auth.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return this.ReuseCloudDeploymentWithProcessSync(token, hubId).
		CreateDeployment(
			jwtToken,
			deployment,
			source,
			optionals)
}

func (this *Controller) RemoveDeployment(token auth.Token, hubId string, deploymentId string) (err error, code int) {
	metadata, err, code := this.processSync.Metadata(token.Jwt(), hubId, deploymentId)
	if err != nil {
		return err, code
	}
	for _, m := range metadata {
		err, code = this.processSync.Remove(token.Jwt(), hubId, m.CamundaDeploymentId)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) SetExecutableFlag(deployment *deploymentmodel.Deployment) {
	this.ReuseCloudDeployment().SetExecutableFlag(deployment)
}

func (this *Controller) StartDeployment(token auth.Token, hubId string, deploymentId string, inputs url.Values) (err error, code int) {
	metadata, err, code := this.processSync.Metadata(token.Jwt(), hubId, deploymentId)
	if err != nil {
		return err, code
	}
	for _, m := range metadata {
		err, code = this.processSync.Start(token.Jwt(), hubId, m.CamundaDeploymentId, inputs)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}
