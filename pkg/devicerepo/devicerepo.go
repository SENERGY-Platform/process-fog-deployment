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
	"github.com/SENERGY-Platform/process-deployment/lib/auth"
	"github.com/SENERGY-Platform/process-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deviceselectionmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
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

func (this *DeviceRepo) GetDeviceGroup(token auth.Token, id string) (result devicemodel.DeviceGroup, err error, code int) {
	return this.reuse.GetDeviceGroup(token, id)
}

func (this *DeviceRepo) GetAspectNode(token auth.Token, id string) (aspectNode devicemodel.AspectNode, err error) {
	return this.reuse.GetAspectNode(token, id)
}

func (this *DeviceRepo) GetDevice(token auth.Token, id string) (devicemodel.Device, error, int) {
	return this.reuse.GetDevice(token, id)
}

func (this *DeviceRepo) GetService(token auth.Token, id string) (devicemodel.Service, error, int) {
	return this.reuse.GetService(token, id)
}

func (this *DeviceRepo) CheckAccess(token auth.Token, kind string, ids []string) (map[string]bool, error) {
	return this.reuse.CheckAccess(token, kind, ids)
}

//deprecated
func (this *DeviceRepo) GetDeviceSelection(token auth.Token, descriptions deviceselectionmodel.FilterCriteriaAndSet, filterByInteraction devicemodel.Interaction) (result []deviceselectionmodel.Selectable, err error, code int) {
	return result, errors.New("should never be called"), http.StatusInternalServerError //it would be possible to call GetBulkDeviceSelection() to get a good result but this method is deprecated and should never be used so new code would be wasted effort
}

type BulkRequestElementWithLocalDeviceFilter struct {
	deviceselectionmodel.BulkRequestElement
	LocalDevices []string `json:"local_devices"`
}

func (this *DeviceRepo) GetBulkDeviceSelection(token auth.Token, bulk deviceselectionmodel.BulkRequest) (result deviceselectionmodel.BulkResult, err error, code int) {
	hub, err, code := this.GetHub(token.Jwt(), this.hubId)
	if err != nil {
		return result, err, code
	}
	bulkWithLocalDevices := []BulkRequestElementWithLocalDeviceFilter{}
	for _, element := range bulk {
		bulkWithLocalDevices = append(bulkWithLocalDevices, BulkRequestElementWithLocalDeviceFilter{
			BulkRequestElement: element,
			LocalDevices:       hub.DeviceLocalIds,
		})
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	buff := new(bytes.Buffer)
	err = json.NewEncoder(buff).Encode(bulkWithLocalDevices)
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
	req.Header.Set("Authorization", token.Jwt())

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

func (this *DeviceRepo) GetHub(token string, id string) (result devicemodel.Hub, err error, code int) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.config.DeviceRepoUrl+"/hubs/"+url.PathEscape(id),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)

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
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}
