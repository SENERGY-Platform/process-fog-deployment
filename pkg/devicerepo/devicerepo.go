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

package devicerepo

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deviceselectionmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

var Factory controller.DeviceRepoFactory = func(config configuration.Config, reuse interfaces.Devices, hubId string) interfaces.Devices {
	return &DeviceRepo{
		config: config,
		hubId:  hubId,
		reuse:  reuse,
	}
}

type DeviceRepo struct {
	config configuration.Config
	hubId  string
	reuse  interfaces.Devices
}

func (this *DeviceRepo) GetDevice(token jwt_http_router.JwtImpersonate, id string) (devicemodel.Device, error, int) {
	return this.reuse.GetDevice(token, id)
}

func (this *DeviceRepo) GetService(token jwt_http_router.JwtImpersonate, id string) (devicemodel.Service, error, int) {
	return this.reuse.GetService(token, id)
}

func (this *DeviceRepo) CheckAccess(token jwt_http_router.JwtImpersonate, kind string, ids []string) (map[string]bool, error) {
	return this.reuse.CheckAccess(token, kind, ids)
}

//TODO: use hub-id in request
func (this *DeviceRepo) GetDeviceSelection(token jwt_http_router.JwtImpersonate, descriptions deviceselectionmodel.FilterCriteriaAndSet, filterByInteraction devicemodel.Interaction) (result []deviceselectionmodel.Selectable, err error, code int) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	payload, err := json.Marshal(descriptions)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	path := "/selectables?json=" + url.QueryEscape(string(payload))
	if filterByInteraction != "" {
		path = path + "&filter_interaction=" + url.QueryEscape(string(filterByInteraction))
	}

	req, err := http.NewRequest(
		"GET",
		this.config.DeviceSelectionUrl+path,
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", string(token))

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, resp.StatusCode
}

//TODO: use hub-id in request
func (this *DeviceRepo) GetBulkDeviceSelection(token jwt_http_router.JwtImpersonate, bulk deviceselectionmodel.BulkRequest) (result deviceselectionmodel.BulkResult, err error, code int) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	buff := new(bytes.Buffer)
	err = json.NewEncoder(buff).Encode(bulk)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	path := "/bulk/selectables"
	req, err := http.NewRequest(
		"POST",
		this.config.DeviceSelectionUrl+path,
		buff,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", string(token))

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err, resp.StatusCode
}
