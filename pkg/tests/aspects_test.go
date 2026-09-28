/*
 * Copyright 2026 InfAI (CC SES)
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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deviceselectionmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/executionmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/api"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/devicerepo"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/processsync"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/tests/mocks"
)

// the fixtures in resources/aspects derive from the ones TestController uses: the same process
// model, once with an aspects list of two nodes in its task payload and once with the deprecated
// single aspect, and a device selection that offers one device carrying both aspects on one path
// option and one device carrying only the first.
const (
	aspectTestHubId             = "urn:infai:ses:hub:114b6d26-5540-44e8-9aeb-234073a49995"
	aspectAir                   = "urn:infai:ses:aspect:air"
	aspectTemperature           = "urn:infai:ses:aspect:temperature"
	deviceWithAirAndTemperature = "urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0060"
	deviceWithAirOnly           = "urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0061"
	aspectTestServiceId         = "urn:infai:ses:service:39415c76-93a3-4e8d-8740-d1a83c64bddc"
	aspectListProcessModelId    = "aspect-list"
	singleAspectProcessModelId  = "aspect-single"
)

func TestAspectList(t *testing.T) {
	env := startAspectTestEnv(t)

	prepared, err := getPreparedDeploymentOf(env.port, aspectListProcessModelId)
	if err != nil {
		t.Fatal(err)
	}
	criteria := prepared.Elements[0].Task.Selection.FilterCriteria

	t.Run("prepared criteria carry the list", func(t *testing.T) {
		if criteria.AspectId != nil {
			t.Errorf("expected no deprecated aspect_id, got %v", *criteria.AspectId)
		}
		if !reflect.DeepEqual(criteria.AspectIds, []string{aspectAir, aspectTemperature}) {
			t.Errorf("unexpected aspect_ids %#v", criteria.AspectIds)
		}
	})

	t.Run("device selection is asked for the list on the hub devices", func(t *testing.T) {
		request := env.selectionRequest(t)
		if !reflect.DeepEqual(request.Criteria[0].AspectIds, []string{aspectAir, aspectTemperature}) {
			t.Errorf("unexpected aspect_ids %#v", request.Criteria[0].AspectIds)
		}
		if request.Criteria[0].AspectId != "" {
			t.Errorf("expected no deprecated aspect_id, got %v", request.Criteria[0].AspectId)
		}
		checkLocalDevices(t, request)
	})

	t.Run("only a device carrying all aspects is offered", func(t *testing.T) {
		offered := offeredDeviceIds(prepared)
		if !reflect.DeepEqual(offered, []string{deviceWithAirAndTemperature}) {
			t.Errorf("unexpected selection options %#v", offered)
		}
	})

	payload := env.deploy(t, prepared, deviceWithAirAndTemperature)

	t.Run("synced deployment keeps the list", func(t *testing.T) {
		synced := payload.deployment.Elements[0].Task.Selection.FilterCriteria
		if !reflect.DeepEqual(synced.AspectIds, []string{aspectAir, aspectTemperature}) {
			t.Errorf("unexpected aspect_ids %#v", synced.AspectIds)
		}
		if synced.AspectId != nil {
			t.Errorf("expected no deprecated aspect_id, got %v", *synced.AspectId)
		}
	})

	t.Run("task payload names the resolved aspect nodes", func(t *testing.T) {
		if payload.task.Aspect != nil {
			t.Errorf("expected no deprecated aspect, got %#v", payload.task.Aspect)
		}
		ids := []string{}
		names := []string{}
		for _, node := range payload.task.Aspects {
			ids = append(ids, node.Id)
			names = append(names, node.Name)
		}
		if !reflect.DeepEqual(ids, []string{aspectAir, aspectTemperature}) {
			t.Errorf("unexpected aspects %#v", ids)
		}
		//the names are only known to the device-repository, so they prove the nodes were resolved there
		if !reflect.DeepEqual(names, []string{"Air", "Temperature"}) {
			t.Errorf("unexpected aspect names %#v", names)
		}
	})
}

// TestSingleAspectIsAliasForList checks the deprecated single aspect: it is passed on in its own
// spelling, and it selects the same devices a list with this one element would.
func TestSingleAspectIsAliasForList(t *testing.T) {
	env := startAspectTestEnv(t)

	prepared, err := getPreparedDeploymentOf(env.port, singleAspectProcessModelId)
	if err != nil {
		t.Fatal(err)
	}
	criteria := prepared.Elements[0].Task.Selection.FilterCriteria

	t.Run("prepared criteria carry the single aspect", func(t *testing.T) {
		if criteria.AspectId == nil || *criteria.AspectId != aspectAir {
			t.Errorf("unexpected aspect_id %#v", criteria.AspectId)
		}
		if len(criteria.AspectIds) != 0 {
			t.Errorf("expected no aspect_ids, got %#v", criteria.AspectIds)
		}
	})

	t.Run("device selection is asked for the single aspect on the hub devices", func(t *testing.T) {
		request := env.selectionRequest(t)
		if request.Criteria[0].AspectId != aspectAir {
			t.Errorf("unexpected aspect_id %#v", request.Criteria[0].AspectId)
		}
		if len(request.Criteria[0].AspectIds) != 0 {
			t.Errorf("expected no aspect_ids, got %#v", request.Criteria[0].AspectIds)
		}
		checkLocalDevices(t, request)
	})

	t.Run("every device carrying the aspect is offered", func(t *testing.T) {
		offered := offeredDeviceIds(prepared)
		if !reflect.DeepEqual(offered, []string{deviceWithAirAndTemperature, deviceWithAirOnly}) {
			t.Errorf("unexpected selection options %#v", offered)
		}
	})

	payload := env.deploy(t, prepared, deviceWithAirOnly)

	t.Run("synced deployment keeps the single aspect", func(t *testing.T) {
		synced := payload.deployment.Elements[0].Task.Selection.FilterCriteria
		if synced.AspectId == nil || *synced.AspectId != aspectAir {
			t.Errorf("unexpected aspect_id %#v", synced.AspectId)
		}
		if len(synced.AspectIds) != 0 {
			t.Errorf("expected no aspect_ids, got %#v", synced.AspectIds)
		}
	})

	t.Run("task payload names the single resolved aspect node", func(t *testing.T) {
		if payload.task.Aspect == nil || payload.task.Aspect.Id != aspectAir || payload.task.Aspect.Name != "Air" {
			t.Errorf("unexpected aspect %#v", payload.task.Aspect)
		}
		if len(payload.task.Aspects) != 0 {
			t.Errorf("expected no aspects, got %#v", payload.task.Aspects)
		}
	})
}

type aspectTestEnv struct {
	port            string
	selectionsCalls *map[string][]string
	syncCalls       *map[string][]string
}

type syncedPayload struct {
	deployment deploymentmodel.Deployment
	task       executionmodel.Task
}

func startAspectTestEnv(t *testing.T) aspectTestEnv {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	permUrl, _ := mocks.NewPermMock(ctx)
	deviceRepoUrl, _, err := mocks.NewStatelessRepoMock(ctx, "resources/aspects/devicerepository.json")
	if err != nil {
		t.Fatal(err)
	}
	syncUrl, syncCalls, err := mocks.NewStatelessRepoMock(ctx, "resources/sync.json")
	if err != nil {
		t.Fatal(err)
	}
	processesUrl, _, err := mocks.NewStatelessRepoMock(ctx, "resources/aspects/processes.json")
	if err != nil {
		t.Fatal(err)
	}
	selectionsUrl, selectionsCalls, err := mocks.NewStatefulRequestMock(ctx, "resources/aspects/selections.json")
	if err != nil {
		t.Fatal(err)
	}
	freePort, err := GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	config := &configuration.ConfigStruct{
		ApiPort:                     strconv.Itoa(freePort),
		DeviceRepoUrl:               deviceRepoUrl,
		ProcessRepoUrl:              processesUrl,
		PermissionsV2Url:            permUrl,
		DeviceSelectionUrl:          selectionsUrl,
		Debug:                       true,
		NotificationUrl:             "http://notification:8080",
		ProcessSyncUrl:              syncUrl,
		EnableDeviceGroupsForTasks:  true,
		EnableDeviceGroupsForEvents: false,
	}
	ctrl, err := controller.New(config, processsync.New(config), devicerepo.Factory)
	if err != nil {
		t.Fatal(err)
	}
	err = api.Start(config, ctx, ctrl)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	return aspectTestEnv{port: config.ApiPort, selectionsCalls: selectionsCalls, syncCalls: syncCalls}
}

// selectionRequest returns the single element of the single bulk request sent to the device-selection.
func (this aspectTestEnv) selectionRequest(t *testing.T) deviceselectionmodel.BulkRequestElementV2 {
	t.Helper()
	calls := (*this.selectionsCalls)["/v2/bulk/selectables"]
	if len(calls) != 1 {
		t.Fatalf("expected one device selection request, got %#v", *this.selectionsCalls)
	}
	request := deviceselectionmodel.BulkRequestV2{}
	err := json.Unmarshal([]byte(calls[0]), &request)
	if err != nil {
		t.Fatal(err)
	}
	if len(request) != 1 || len(request[0].Criteria) != 1 {
		t.Fatalf("unexpected device selection request %v", calls[0])
	}
	return request[0]
}

// deploy selects the device for the task, sends the deployment and returns what reached the process-sync.
func (this aspectTestEnv) deploy(t *testing.T, prepared deploymentmodel.Deployment, deviceId string) (result syncedPayload) {
	t.Helper()
	serviceId := aspectTestServiceId
	prepared.Elements[0].Task.Selection.SelectedDeviceId = &deviceId
	prepared.Elements[0].Task.Selection.SelectedServiceId = &serviceId
	err := getTestSendDeployment(this.port, prepared)
	if err != nil {
		t.Fatal(err)
	}
	calls := (*this.syncCalls)["/deployments/"+aspectTestHubId]
	if len(calls) != 1 {
		t.Fatalf("expected one process-sync request, got %#v", *this.syncCalls)
	}
	err = json.Unmarshal([]byte(calls[0]), &result.deployment)
	if err != nil {
		t.Fatal(err)
	}
	result.task, err = deployedTaskPayload(result.deployment.Diagram.XmlDeployed)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

var cdataPattern = regexp.MustCompile(`(?s)<!\[CDATA\[(.*?)]]>`)

// deployedTaskPayload reads the camunda task payload out of a deployed bpmn with a single task.
// The payload is the only input parameter the stringifier writes as CDATA.
func deployedTaskPayload(xml string) (result executionmodel.Task, err error) {
	matches := cdataPattern.FindAllStringSubmatch(xml, -1)
	if len(matches) != 1 {
		return result, fmt.Errorf("expected one task payload in deployed xml, found %v", len(matches))
	}
	err = json.Unmarshal([]byte(matches[0][1]), &result)
	return result, err
}

func checkLocalDevices(t *testing.T, request deviceselectionmodel.BulkRequestElementV2) {
	t.Helper()
	//the local ids of resources/aspects/devicerepository.json hub
	expected := []string{"e3a7a0a7f35c9c9615839eca59db5b7d-43", "2"}
	if !reflect.DeepEqual(request.LocalDevices, expected) {
		t.Errorf("expected the local devices of the hub %#v, got %#v", expected, request.LocalDevices)
	}
}

// offeredDeviceIds returns the devices with at least one service matching the criteria. A device
// whose services all miss the criteria stays in the options, but without a service to select.
func offeredDeviceIds(deployment deploymentmodel.Deployment) (result []string) {
	result = []string{}
	for _, option := range deployment.Elements[0].Task.Selection.SelectionOptions {
		if option.Device != nil && len(option.Services) > 0 {
			result = append(result, option.Device.Id)
		}
	}
	return result
}

func getPreparedDeploymentOf(port string, modelId string) (result deploymentmodel.Deployment, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", "http://localhost:"+port+"/prepared-deployments/"+aspectTestHubId+"/"+modelId, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		temp, _ := io.ReadAll(resp.Body)
		return result, errors.New(fmt.Sprint(resp.StatusCode, string(temp)))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	if len(result.Elements) != 1 || result.Elements[0].Task == nil {
		return result, fmt.Errorf("expected a single task element, got %#v", result.Elements)
	}
	return result, nil
}
