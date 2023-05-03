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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/process-deployment/lib/config"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/devicemodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/api"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/devicerepo"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/processsync"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/tests/docker"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/tests/mocks"
	paho "github.com/eclipse/paho.mqtt.golang"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.NewId = func() string {
		return "test-id"
	}

	deviceManagerUrl, deviceRepoUrl, permsearchUrl, err := docker.DeviceManagerWithDependencies(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	_, selectionIp, err := docker.DeviceSelection(ctx, wg, "", deviceRepoUrl, permsearchUrl)
	if err != nil {
		t.Error(err)
		return
	}

	_, mqttip, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	mqttBroker := "tcp://" + mqttip + ":1883"

	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	mongoUrl := "mongodb://" + mongoIp + ":27017"

	_, syncIp, err := docker.ProcessSync(ctx, wg, permsearchUrl, mqttBroker, mongoUrl)
	if err != nil {
		t.Error(err)
		return
	}
	syncUrl := "http://" + syncIp + ":8080"

	processesUrl, _, err := mocks.NewStatelessRepoMock(ctx, "resources/processes.json")
	if err != nil {
		t.Error(err)
		return
	}

	var camundaPgIp string
	camundaDb, camundaPgIp, _, err := docker.Postgres(ctx, wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, wg, camundaPgIp, "5432")
	if err != nil {
		t.Error(err)
		return
	}

	_, clientMongoIp, err := docker.MongoDB(ctx, wg)
	clientMetadataStorageUrl := "mongodb://" + clientMongoIp + ":27017/metadata"

	err = docker.MgwProcessSyncClient(ctx, wg, camundaDb, camundaUrl, mqttBroker, "mgw-test-sync-client", "urn:infai:ses:hubs:h1", clientMetadataStorageUrl)
	if err != nil {
		t.Error(err)
		return
	}

	freePort, err := GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config := &configuration.ConfigStruct{
		ApiPort:                     strconv.Itoa(freePort),
		DeviceRepoUrl:               deviceRepoUrl,
		ProcessRepoUrl:              processesUrl,
		PermSearchUrl:               permsearchUrl,
		DeviceSelectionUrl:          "http://" + selectionIp + ":8080",
		Debug:                       true,
		NotificationUrl:             "http://notification:8080",
		ProcessSyncUrl:              syncUrl,
		EnableDeviceGroupsForTasks:  true,
		EnableDeviceGroupsForEvents: false,
	}

	ctrl, err := controller.New(config, processsync.New(config), devicerepo.Factory)
	if err != nil {
		t.Error(err)
		return
	}

	err = api.Start(config, ctx, ctrl)
	if err != nil {
		t.Error(err)
		return
	}

	mqtt := paho.NewClient(paho.NewClientOptions().AddBroker(mqttBroker).SetAutoReconnect(true))
	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
		t.Error("Error on Mqtt.Connect(): ", token.Error())
		return
	}
	messages := map[string][]string{}
	msgMux := sync.Mutex{}
	token := mqtt.Subscribe("#", 2, func(client paho.Client, message paho.Message) {
		msgMux.Lock()
		defer msgMux.Unlock()
		messages[message.Topic()] = append(messages[message.Topic()], string(message.Payload()))
	})
	if token.Wait() && token.Error() != nil {
		t.Error("Error on Mqtt.Subscribe(): ", token.Error())
		return
	}

	aspects := []devicemodel.Aspect{{
		Id:   "urn:infai:ses:aspect:a1",
		Name: "a1",
	}}

	characteristics := []devicemodel.Characteristic{{
		Id:   "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
		Name: "characteristic",
		Type: devicemodel.String,
	}}

	concepts := []devicemodel.Concept{{
		Id:                   "urn:infai:ses:concept:0bc81398-3ed6-4e2b-a6c4-b754583aac37",
		Name:                 "concept",
		CharacteristicIds:    []string{"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"},
		BaseCharacteristicId: "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
	}}

	functions := []devicemodel.Function{{
		Id:        "urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c",
		Name:      "setTemperatureFunction",
		ConceptId: "urn:infai:ses:concept:0bc81398-3ed6-4e2b-a6c4-b754583aac37",
	}}

	deviceClasses := []devicemodel.DeviceClass{{
		Id:   "urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393",
		Name: "dc",
	}}

	protocol := []devicemodel.Protocol{{
		Id:      "urn:infai:ses:protocol:p1",
		Name:    "p1",
		Handler: "p1",
		ProtocolSegments: []devicemodel.ProtocolSegment{
			{
				Id:   "ps1",
				Name: "ps1",
			},
		},
	}}

	deviceTypes := []devicemodel.DeviceType{{
		Id:   "urn:infai:ses:device-type:dt1",
		Name: "dt",
		Services: []devicemodel.Service{{
			Id:          "urn:infai:ses:service:s1",
			LocalId:     "s1",
			Name:        "s1",
			Interaction: devicemodel.REQUEST,
			ProtocolId:  "urn:infai:ses:protocol:p1",
			Inputs: []devicemodel.Content{{
				Id: "content1",
				ContentVariable: devicemodel.ContentVariable{
					Id:               "cv1",
					Name:             "foo",
					Type:             devicemodel.String,
					CharacteristicId: "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
					FunctionId:       "urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c",
				},
				Serialization:     "json",
				ProtocolSegmentId: "ps1",
			}},
		}},
		DeviceClassId: "urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393",
	}}

	devices := []devicemodel.Device{{
		Id:           "urn:infai:ses:device:d1",
		LocalId:      "d1",
		Name:         "d1",
		DeviceTypeId: "urn:infai:ses:device-type:dt1",
	}}

	hubs := []devicemodel.Hub{{
		Id:             "urn:infai:ses:hubs:h1",
		Name:           "h1",
		Hash:           "",
		DeviceLocalIds: []string{"d1"},
	}}

	t.Run("create characteristics", func(t *testing.T) {
		for _, c := range characteristics {
			err = setCharacteristic(deviceManagerUrl, c)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("create concepts", func(t *testing.T) {
		for _, c := range concepts {
			err = setConcept(deviceManagerUrl, c)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("create aspects", func(t *testing.T) {
		for _, a := range aspects {
			err = setAspect(deviceManagerUrl, a)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("create functions", func(t *testing.T) {
		for _, f := range functions {
			err = setFunction(deviceManagerUrl, f)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("create device-classes", func(t *testing.T) {
		for _, dc := range deviceClasses {
			err = setDeviceClass(deviceManagerUrl, dc)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("create protocols", func(t *testing.T) {
		for _, p := range protocol {
			err = setProtocol(deviceManagerUrl, p)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("create device-types", func(t *testing.T) {
		for _, dt := range deviceTypes {
			err = setDeviceType(deviceManagerUrl, dt)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("create devices", func(t *testing.T) {
		for _, d := range devices {
			err = setDevice(deviceManagerUrl, d)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("create hubs", func(t *testing.T) {
		for _, h := range hubs {
			err = setHub(deviceManagerUrl, h)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	var depl deploymentmodel.Deployment
	t.Run("prepare", func(t *testing.T) {
		depl, err = Jwtget[deploymentmodel.Deployment](AdminJwt, "http://localhost:"+strconv.Itoa(freePort)+"/prepared-deployments/"+url.PathEscape("urn:infai:ses:hubs:h1")+"/e32329bc-3800-4429-986e-4cc208e95fc2")
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("deploy", func(t *testing.T) {
		depl.Elements[0].Task.Selection.SelectedDeviceId = &depl.Elements[0].Task.Selection.SelectionOptions[0].Device.Id
		depl.Elements[0].Task.Selection.SelectedServiceId = &depl.Elements[0].Task.Selection.SelectionOptions[0].Services[0].Id
		depl.Elements[0].Task.Selection.SelectedPath = &depl.Elements[0].Task.Selection.SelectionOptions[0].PathOptions[depl.Elements[0].Task.Selection.SelectionOptions[0].Services[0].Id][0]
		resp, err := Jwtpost(AdminJwt, "http://localhost:"+strconv.Itoa(freePort)+"/deployments/"+url.PathEscape("urn:infai:ses:hubs:h1"), depl)
		if err != nil {
			t.Error(err)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&depl)
		if err != nil {
			t.Error(err)
			return
		}
	})

	time.Sleep(2 * time.Second)

	t.Run("check deployment sync after deploy", func(t *testing.T) {
		result, err := Jwtget[[]CamundaDeployment](AdminJwt, config.ProcessSyncUrl+"/deployments"+"?network_id="+url.QueryEscape("urn:infai:ses:hubs:h1"))
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) != 1 || result[0].IsPlaceholder || result[0].MarkedForDelete {
			t.Error(result)
			return
		}
	})

	t.Run("start", func(t *testing.T) {
		_, err = Jwtget[bool](AdminJwt, "http://localhost:"+strconv.Itoa(freePort)+"/deployments/"+url.PathEscape("urn:infai:ses:hubs:h1")+"/"+url.PathEscape(depl.Id)+"/start")
		if err != nil {
			t.Error(err)
			return
		}
	})

	time.Sleep(2 * time.Second)

	t.Run("check instance sync after start", func(t *testing.T) {
		result, err := Jwtget[[]ProcessInstance](AdminJwt, config.ProcessSyncUrl+"/process-instances"+"?network_id="+url.QueryEscape("urn:infai:ses:hubs:h1"))
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) != 1 {
			t.Error(result)
			return
		}
	})

	t.Run("delete", func(t *testing.T) {
		_, err = Jwtdelete[bool](AdminJwt, "http://localhost:"+strconv.Itoa(freePort)+"/deployments/"+url.PathEscape("urn:infai:ses:hubs:h1")+"/"+url.PathEscape(depl.Id))
		if err != nil {
			t.Error(err)
			return
		}
	})

	time.Sleep(2 * time.Second)

	t.Run("check deployment sync after delete", func(t *testing.T) {
		result, err := Jwtget[[]CamundaDeployment](AdminJwt, config.ProcessSyncUrl+"/deployments?network_id="+url.QueryEscape("urn:infai:ses:hubs:h1"))
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) != 0 {
			t.Error(result)
			return
		}
	})
}

func setAspect(devicemanagerUrl string, a devicemodel.Aspect) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/aspects/"+url.PathEscape(a.Id), a)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setCharacteristic(devicemanagerUrl string, c devicemodel.Characteristic) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/characteristics/"+url.PathEscape(c.Id), c)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setConcept(devicemanagerUrl string, c devicemodel.Concept) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/concepts/"+url.PathEscape(c.Id), c)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setFunction(devicemanagerUrl string, f devicemodel.Function) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/functions/"+url.PathEscape(f.Id), f)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setDeviceClass(devicemanagerUrl string, dc devicemodel.DeviceClass) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/device-classes/"+url.PathEscape(dc.Id), dc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setProtocol(devicemanagerUrl string, p devicemodel.Protocol) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/protocols/"+url.PathEscape(p.Id), p)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setDeviceType(devicemanagerUrl string, dt devicemodel.DeviceType) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/device-types/"+url.PathEscape(dt.Id), dt)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setDevice(devicemanagerUrl string, d devicemodel.Device) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/devices/"+url.PathEscape(d.Id), d)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func setHub(devicemanagerUrl string, h devicemodel.Hub) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/hubs/"+url.PathEscape(h.Id), h)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

var SleepAfterEdit = 2 * time.Second

func Jwtput(token string, url string, msg interface{}) (resp *http.Response, err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(msg)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if SleepAfterEdit != 0 {
		time.Sleep(SleepAfterEdit)
	}
	return
}

func Jwtpost(token string, url string, msg interface{}) (resp *http.Response, err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(msg)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if SleepAfterEdit != 0 {
		time.Sleep(SleepAfterEdit)
	}
	return
}

func Jwtget[T any](token string, url string) (result T, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("Jwtget %v: \n%v\n%v", url, resp.StatusCode, string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func Jwtdelete[T any](token string, url string) (result T, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("Jwtget %v: \n%v\n%v", url, resp.StatusCode, string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

const AdminJwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"

type HistoricProcessInstance struct {
	Id                       string  `json:"id"`
	SuperProcessInstanceId   string  `json:"superProcessInstanceId"`
	SuperCaseInstanceId      string  `json:"superCaseInstanceId"`
	CaseInstanceId           string  `json:"caseInstanceId"`
	ProcessDefinitionName    string  `json:"processDefinitionName"`
	ProcessDefinitionKey     string  `json:"processDefinitionKey"`
	ProcessDefinitionVersion float64 `json:"processDefinitionVersion"`
	ProcessDefinitionId      string  `json:"processDefinitionId"`
	BusinessKey              string  `json:"businessKey"`
	StartTime                string  `json:"startTime"`
	EndTime                  string  `json:"endTime"`
	DurationInMillis         float64 `json:"durationInMillis"`
	StartUserId              string  `json:"startUserId"`
	StartActivityId          string  `json:"startActivityId"`
	DeleteReason             string  `json:"deleteReason"`
	TenantId                 string  `json:"tenantId"`
	State                    string  `json:"state"`
}

type ProcessInstance struct {
	Id             string `json:"id,omitempty"`
	DefinitionId   string `json:"definitionId,omitempty"`
	BusinessKey    string `json:"businessKey,omitempty"`
	CaseInstanceId string `json:"caseInstanceId,omitempty"`
	Ended          bool   `json:"ended,omitempty"`
	Suspended      bool   `json:"suspended,omitempty"`
	TenantId       string `json:"tenantId,omitempty"`
}

type CamundaDeployment struct {
	camundaDeployment
	SyncInfo
}

type SyncInfo struct {
	NetworkId       string    `json:"network_id"`
	IsPlaceholder   bool      `json:"is_placeholder"`
	MarkedForDelete bool      `json:"marked_for_delete"`
	SyncDate        time.Time `json:"sync_date"`
}

type camundaDeployment struct {
	Id             string      `json:"id"`
	Name           string      `json:"name"`
	Source         string      `json:"source"`
	DeploymentTime interface{} `json:"deploymentTime"`
	TenantId       string      `json:"tenantId"`
}
