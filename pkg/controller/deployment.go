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
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
)

func (this *Controller) PrepareDeployment(token string, hubId string, xml string, svg string) (result deploymentmodel.Deployment, err error, code int) {
	result, err = this.deploymentParser.PrepareDeployment(xml)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	err = this.ReuseCloudDeploymentWithNewDeviceRepo(hubId).SetDeploymentOptionsV2(jwt_http_router.JwtImpersonate(token), &result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result.Diagram.Svg = svg
	this.SetExecutableFlag(&result)
	return result, nil, http.StatusOK
}

func (this *Controller) CreateDeployment(token string, hubId string, deployment deploymentmodel.Deployment, source string) (err error, code int) {
	jwtToken := jwt_http_router.Jwt{}
	err = jwt_http_router.GetJWTPayload(token, &jwtToken)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	_, err, code = this.ReuseCloudDeploymentWithProcessSync(token, hubId).
		CreateDeploymentV2(
			jwtToken,
			deployment,
			source)
	return
}

func (this *Controller) SetExecutableFlag(deployment *deploymentmodel.Deployment) {
	this.ReuseCloudDeployment().SetExecutableFlagV2(deployment)
}
