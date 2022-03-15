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

package processsync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"net/http"
	"net/url"
	"time"
)

func New(config configuration.Config) *ProcessSync {
	return &ProcessSync{
		config: config,
	}
}

type ProcessSync struct {
	config configuration.Config
}

func (this *ProcessSync) Deploy(token string, hubId string, deployment deploymentmodel.Deployment) error {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(deployment)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", this.config.ProcessSyncUrl+"/deployments/"+url.PathEscape(hubId), requestBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", token)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
	}
	return err
}
