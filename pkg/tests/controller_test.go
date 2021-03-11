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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/api"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/devicerepo"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/processsync"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/tests/mocks"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	permsearchUrl, permsearchCalls := mocks.NewPermSearchMock(ctx)
	deviceRepoUrl, deviceRepoCalls, err := mocks.NewStatelessRepoMock(ctx, "resources/devicerepository.json")
	if err != nil {
		t.Error(err)
		return
	}
	syncUrl, syncCalls, err := mocks.NewStatelessRepoMock(ctx, "resources/sync.json")
	if err != nil {
		t.Error(err)
		return
	}
	processesUrl, processesCalls, err := mocks.NewStatelessRepoMock(ctx, "resources/processes.json")
	if err != nil {
		t.Error(err)
		return
	}
	selectionsUrl, selectionsCalls, err := mocks.NewStatefulRequestMock(ctx, "resources/selections.json")
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
		DeviceSelectionUrl:          selectionsUrl,
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

	prepared, err := getTestPreparedDeployment(config.ApiPort)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(prepared, deploymentmodel.Deployment{
		Name: "set_target_temp",
		Diagram: deploymentmodel.Diagram{
			XmlRaw:      "\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbpmn:definitions xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:camunda=\"http://camunda.org/schema/1.0/bpmn\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_1\" targetNamespace=\"http://bpmn.io/schema/bpmn\"\u003e\u003cbpmn:process id=\"set_target_temp\" isExecutable=\"true\"\u003e\u003cbpmn:startEvent id=\"StartEvent_1\"\u003e\u003cbpmn:outgoing\u003eSequenceFlow_058cir1\u003c/bpmn:outgoing\u003e\u003c/bpmn:startEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_058cir1\" sourceRef=\"StartEvent_1\" targetRef=\"Task_18tgni4\" /\u003e\u003cbpmn:endEvent id=\"EndEvent_1uie5qv\"\u003e\u003cbpmn:incoming\u003eSequenceFlow_1hory47\u003c/bpmn:incoming\u003e\u003c/bpmn:endEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_1hory47\" sourceRef=\"Task_18tgni4\" targetRef=\"EndEvent_1uie5qv\" /\u003e\u003cbpmn:serviceTask id=\"Task_18tgni4\" name=\"Thermostat setTemperatureFunction\" camunda:type=\"external\" camunda:topic=\"pessimistic\"\u003e\u003cbpmn:extensionElements\u003e\u003ccamunda:inputOutput\u003e\u003ccamunda:inputParameter name=\"payload\"\u003e{\n    \"function\": {\n        \"id\": \"urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c\",\n        \"name\": \"setTemperatureFunction\",\n        \"concept_id\": \"urn:infai:ses:concept:0bc81398-3ed6-4e2b-a6c4-b754583aac37\",\n        \"rdf_type\": \"https://senergy.infai.org/ontology/ControllingFunction\"\n    },\n    \"device_class\": {\n        \"id\": \"urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393\",\n        \"name\": \"Thermostat\",\n        \"rdf_type\": \"https://senergy.infai.org/ontology/DeviceClass\"\n    },\n    \"aspect\": null,\n    \"label\": \"setTemperatureFunction\",\n    \"input\": 0,\n    \"characteristic_id\": \"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a\",\n    \"retries\": 0\n}\u003c/camunda:inputParameter\u003e\u003ccamunda:inputParameter name=\"inputs\"\u003e0\u003c/camunda:inputParameter\u003e\u003c/camunda:inputOutput\u003e\u003c/bpmn:extensionElements\u003e\u003cbpmn:incoming\u003eSequenceFlow_058cir1\u003c/bpmn:incoming\u003e\u003cbpmn:outgoing\u003eSequenceFlow_1hory47\u003c/bpmn:outgoing\u003e\u003c/bpmn:serviceTask\u003e\u003c/bpmn:process\u003e\u003cbpmndi:BPMNDiagram id=\"BPMNDiagram_1\"\u003e\u003cbpmndi:BPMNPlane id=\"BPMNPlane_1\" bpmnElement=\"set_target_temp\"\u003e\u003cbpmndi:BPMNShape id=\"_BPMNShape_StartEvent_2\" bpmnElement=\"StartEvent_1\"\u003e\u003cdc:Bounds x=\"173\" y=\"102\" width=\"36\" height=\"36\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_058cir1_di\" bpmnElement=\"SequenceFlow_058cir1\"\u003e\u003cdi:waypoint x=\"209\" y=\"120\" /\u003e\u003cdi:waypoint x=\"260\" y=\"120\" /\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"EndEvent_1uie5qv_di\" bpmnElement=\"EndEvent_1uie5qv\"\u003e\u003cdc:Bounds x=\"412\" y=\"102\" width=\"36\" height=\"36\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_1hory47_di\" bpmnElement=\"SequenceFlow_1hory47\"\u003e\u003cdi:waypoint x=\"360\" y=\"120\" /\u003e\u003cdi:waypoint x=\"412\" y=\"120\" /\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"ServiceTask_1r7hcop_di\" bpmnElement=\"Task_18tgni4\"\u003e\u003cdc:Bounds x=\"260\" y=\"80\" width=\"100\" height=\"80\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003c/bpmndi:BPMNPlane\u003e\u003c/bpmndi:BPMNDiagram\u003e\u003c/bpmn:definitions\u003e",
			XmlDeployed: "",
			Svg:         "\u003c?xml version=\"1.0\" encoding=\"utf-8\"?\u003e\n\u003c!-- created with bpmn-js / http://bpmn.io --\u003e\n\u003c!DOCTYPE svg PUBLIC \"-//W3C//DTD SVG 1.1//EN\" \"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd\"\u003e\n\u003csvg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" width=\"287\" height=\"92\" viewBox=\"167 74 287 92\" version=\"1.1\"\u003e\u003cdefs\u003e\u003cmarker id=\"sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew\" viewBox=\"0 0 20 20\" refX=\"11\" refY=\"10\" markerWidth=\"10\" markerHeight=\"10\" orient=\"auto\"\u003e\u003cpath d=\"M 1 5 L 11 10 L 1 15 Z\" style=\"fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;\"/\u003e\u003c/marker\u003e\u003c/defs\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-connection\" data-element-id=\"SequenceFlow_058cir1\" style=\"display: block;\"\u003e\u003cg class=\"djs-visual\"\u003e\u003cpath d=\"m  209,120L260,120 \" style=\"fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url('#sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew');\"/\u003e\u003c/g\u003e\u003cpolyline points=\"209,120 260,120 \" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"203\" y=\"114\" width=\"63\" height=\"12\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-connection\" data-element-id=\"SequenceFlow_1hory47\" style=\"display: block;\"\u003e\u003cg class=\"djs-visual\"\u003e\u003cpath d=\"m  360,120L412,120 \" style=\"fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url('#sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew');\"/\u003e\u003c/g\u003e\u003cpolyline points=\"360,120 412,120 \" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"354\" y=\"114\" width=\"64\" height=\"12\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"StartEvent_1\" style=\"display: block;\" transform=\"matrix(1 0 0 1 173 102)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003ccircle cx=\"18\" cy=\"18\" r=\"18\" style=\"stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"36\" height=\"36\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"48\" height=\"48\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"EndEvent_1uie5qv\" style=\"display: block;\" transform=\"matrix(1 0 0 1 412 102)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003ccircle cx=\"18\" cy=\"18\" r=\"18\" style=\"stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"36\" height=\"36\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"48\" height=\"48\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"Task_18tgni4\" style=\"display: block;\" transform=\"matrix(1 0 0 1 260 80)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003crect x=\"0\" y=\"0\" width=\"100\" height=\"80\" rx=\"10\" ry=\"10\" style=\"stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;\"/\u003e\u003ctext lineHeight=\"1.2\" class=\"djs-label\" style=\"font-family: Arial, sans-serif; font-size: 12px; font-weight: normal; fill: black;\"\u003e\u003ctspan x=\"19.3203125\" y=\"29.200000000000003\"\u003eThermostat \u003c/tspan\u003e\u003ctspan x=\"8.1484375\" y=\"43.6\"\u003esetTemperature\u003c/tspan\u003e\u003ctspan x=\"27.3203125\" y=\"58\"\u003eFunction\u003c/tspan\u003e\u003c/text\u003e\u003cpath d=\"m 12,18 v -1.71335 c 0.352326,-0.0705 0.703932,-0.17838 1.047628,-0.32133 0.344416,-0.14465 0.665822,-0.32133 0.966377,-0.52145 l 1.19431,1.18005 1.567487,-1.57688 -1.195028,-1.18014 c 0.403376,-0.61394 0.683079,-1.29908 0.825447,-2.01824 l 1.622133,-0.01 v -2.2196 l -1.636514,0.01 c -0.07333,-0.35153 -0.178319,-0.70024 -0.323564,-1.04372 -0.145244,-0.34406 -0.321407,-0.6644 -0.522735,-0.96217 l 1.131035,-1.13631 -1.583305,-1.56293 -1.129598,1.13589 c -0.614052,-0.40108 -1.302883,-0.68093 -2.022633,-0.82247 l 0.0093,-1.61852 h -2.241173 l 0.0042,1.63124 c -0.353763,0.0736 -0.705369,0.17977 -1.049785,0.32371 -0.344415,0.14437 -0.665102,0.32092 -0.9635006,0.52046 l -1.1698628,-1.15823 -1.5667691,1.5792 1.1684265,1.15669 c -0.4026573,0.61283 -0.68308,1.29797 -0.8247287,2.01713 l -1.6588041,0.003 v 2.22174 l 1.6724648,-0.006 c 0.073327,0.35077 0.1797598,0.70243 0.3242851,1.04472 0.1452428,0.34448 0.3214064,0.6644 0.5227339,0.96066 l -1.1993431,1.19723 1.5840256,1.56011 1.1964668,-1.19348 c 0.6140517,0.40346 1.3028827,0.68232 2.0233517,0.82331 l 7.19e-4,1.69892 h 2.226848 z m 0.221462,-3.9957 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 1px; stroke: black;\"/\u003e\u003cpath d=\"m 17.2,18 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 0px; stroke: black;\"/\u003e\u003cpath d=\"m 17,22 v -1.71335 c 0.352326,-0.0705 0.703932,-0.17838 1.047628,-0.32133 0.344416,-0.14465 0.665822,-0.32133 0.966377,-0.52145 l 1.19431,1.18005 1.567487,-1.57688 -1.195028,-1.18014 c 0.403376,-0.61394 0.683079,-1.29908 0.825447,-2.01824 l 1.622133,-0.01 v -2.2196 l -1.636514,0.01 c -0.07333,-0.35153 -0.178319,-0.70024 -0.323564,-1.04372 -0.145244,-0.34406 -0.321407,-0.6644 -0.522735,-0.96217 l 1.131035,-1.13631 -1.583305,-1.56293 -1.129598,1.13589 c -0.614052,-0.40108 -1.302883,-0.68093 -2.022633,-0.82247 l 0.0093,-1.61852 h -2.241173 l 0.0042,1.63124 c -0.353763,0.0736 -0.705369,0.17977 -1.049785,0.32371 -0.344415,0.14437 -0.665102,0.32092 -0.9635006,0.52046 l -1.1698628,-1.15823 -1.5667691,1.5792 1.1684265,1.15669 c -0.4026573,0.61283 -0.68308,1.29797 -0.8247287,2.01713 l -1.6588041,0.003 v 2.22174 l 1.6724648,-0.006 c 0.073327,0.35077 0.1797598,0.70243 0.3242851,1.04472 0.1452428,0.34448 0.3214064,0.6644 0.5227339,0.96066 l -1.1993431,1.19723 1.5840256,1.56011 1.1964668,-1.19348 c 0.6140517,0.40346 1.3028827,0.68232 2.0233517,0.82331 l 7.19e-4,1.69892 h 2.226848 z m 0.221462,-3.9957 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 1px; stroke: black;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"100\" height=\"80\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"112\" height=\"92\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003c/svg\u003e",
		},
		Elements: []deploymentmodel.Element{
			{
				BpmnId: "Task_18tgni4",
				Name:   "Thermostat setTemperatureFunction",
				Task: &deploymentmodel.Task{
					Parameter: map[string]string{"inputs": "0"},
					Selection: deploymentmodel.Selection{
						FilterCriteria: deploymentmodel.FilterCriteria{
							CharacteristicId: strptr("urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a"),
							FunctionId:       strptr("urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c"),
							DeviceClassId:    strptr("urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393"),
						},
					},
				},
			},
		},
		Executable: false,
	}) {
		temp, _ := json.Marshal(prepared)
		t.Error(string(temp))
		return
	}

	deviceId := "urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0060"
	serviceId := "urn:infai:ses:service:39415c76-93a3-4e8d-8740-d1a83c64bddc"
	prepared.Elements[0].Task.Selection.SelectedDeviceId = &deviceId
	prepared.Elements[0].Task.Selection.SelectedServiceId = &serviceId

	err = getTestSendDeployment(config.ApiPort, prepared)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(*deviceRepoCalls, map[string][]string{"/hubs/urn:infai:ses:hub:114b6d26-5540-44e8-9aeb-234073a49995": {""}}) {
		t.Error(*deviceRepoCalls)
	}
	if !reflect.DeepEqual(*processesCalls, map[string][]string{"/processes/e32329bc-3800-4429-986e-4cc208e95fc2": {""}}) {
		t.Error(*processesCalls)
	}
	if !reflect.DeepEqual(*permsearchCalls, map[string][]string{"/v2/query": {
		`{"resource":"devices","find":null,"list_ids":null,"check_ids":{"ids":["urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0060"],"rights":"x"}}`,
		`{"resource":"device-groups","find":null,"list_ids":null,"check_ids":{"ids":[],"rights":"x"}}`}}) {
		t.Error(*permsearchCalls)
	}
	if !reflect.DeepEqual(*selectionsCalls, map[string][]string{"/bulk/selectables": {`[{"id":"Task_18tgni4","filter_interaction":"event","filter_protocols":null,"criteria":[{"function_id":"urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c","device_class_id":"urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393","aspect_id":""}],"include_groups":true,"local_devices":["e3a7a0a7f35c9c9615839eca59db5b7d-43","2"]}]`}}) {
		t.Error(*selectionsCalls)
	}
	if !reflect.DeepEqual(*syncCalls, map[string][]string{"/deployments/urn:infai:ses:hub:114b6d26-5540-44e8-9aeb-234073a49995": {`{"id":"unused-id","name":"set_target_temp","description":"","diagram":{"xml_raw":"\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbpmn:definitions xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:camunda=\"http://camunda.org/schema/1.0/bpmn\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_1\" targetNamespace=\"http://bpmn.io/schema/bpmn\"\u003e\u003cbpmn:process id=\"set_target_temp\" isExecutable=\"true\"\u003e\u003cbpmn:startEvent id=\"StartEvent_1\"\u003e\u003cbpmn:outgoing\u003eSequenceFlow_058cir1\u003c/bpmn:outgoing\u003e\u003c/bpmn:startEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_058cir1\" sourceRef=\"StartEvent_1\" targetRef=\"Task_18tgni4\" /\u003e\u003cbpmn:endEvent id=\"EndEvent_1uie5qv\"\u003e\u003cbpmn:incoming\u003eSequenceFlow_1hory47\u003c/bpmn:incoming\u003e\u003c/bpmn:endEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_1hory47\" sourceRef=\"Task_18tgni4\" targetRef=\"EndEvent_1uie5qv\" /\u003e\u003cbpmn:serviceTask id=\"Task_18tgni4\" name=\"Thermostat setTemperatureFunction\" camunda:type=\"external\" camunda:topic=\"pessimistic\"\u003e\u003cbpmn:extensionElements\u003e\u003ccamunda:inputOutput\u003e\u003ccamunda:inputParameter name=\"payload\"\u003e{\n    \"function\": {\n        \"id\": \"urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c\",\n        \"name\": \"setTemperatureFunction\",\n        \"concept_id\": \"urn:infai:ses:concept:0bc81398-3ed6-4e2b-a6c4-b754583aac37\",\n        \"rdf_type\": \"https://senergy.infai.org/ontology/ControllingFunction\"\n    },\n    \"device_class\": {\n        \"id\": \"urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393\",\n        \"name\": \"Thermostat\",\n        \"rdf_type\": \"https://senergy.infai.org/ontology/DeviceClass\"\n    },\n    \"aspect\": null,\n    \"label\": \"setTemperatureFunction\",\n    \"input\": 0,\n    \"characteristic_id\": \"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a\",\n    \"retries\": 0\n}\u003c/camunda:inputParameter\u003e\u003ccamunda:inputParameter name=\"inputs\"\u003e0\u003c/camunda:inputParameter\u003e\u003c/camunda:inputOutput\u003e\u003c/bpmn:extensionElements\u003e\u003cbpmn:incoming\u003eSequenceFlow_058cir1\u003c/bpmn:incoming\u003e\u003cbpmn:outgoing\u003eSequenceFlow_1hory47\u003c/bpmn:outgoing\u003e\u003c/bpmn:serviceTask\u003e\u003c/bpmn:process\u003e\u003cbpmndi:BPMNDiagram id=\"BPMNDiagram_1\"\u003e\u003cbpmndi:BPMNPlane id=\"BPMNPlane_1\" bpmnElement=\"set_target_temp\"\u003e\u003cbpmndi:BPMNShape id=\"_BPMNShape_StartEvent_2\" bpmnElement=\"StartEvent_1\"\u003e\u003cdc:Bounds x=\"173\" y=\"102\" width=\"36\" height=\"36\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_058cir1_di\" bpmnElement=\"SequenceFlow_058cir1\"\u003e\u003cdi:waypoint x=\"209\" y=\"120\" /\u003e\u003cdi:waypoint x=\"260\" y=\"120\" /\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"EndEvent_1uie5qv_di\" bpmnElement=\"EndEvent_1uie5qv\"\u003e\u003cdc:Bounds x=\"412\" y=\"102\" width=\"36\" height=\"36\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_1hory47_di\" bpmnElement=\"SequenceFlow_1hory47\"\u003e\u003cdi:waypoint x=\"360\" y=\"120\" /\u003e\u003cdi:waypoint x=\"412\" y=\"120\" /\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"ServiceTask_1r7hcop_di\" bpmnElement=\"Task_18tgni4\"\u003e\u003cdc:Bounds x=\"260\" y=\"80\" width=\"100\" height=\"80\" /\u003e\u003c/bpmndi:BPMNShape\u003e\u003c/bpmndi:BPMNPlane\u003e\u003c/bpmndi:BPMNDiagram\u003e\u003c/bpmn:definitions\u003e","xml_deployed":"\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbpmn:definitions xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xmlns:bpmn=\"http://www.omg.org/spec/BPMN/20100524/MODEL\" xmlns:bpmndi=\"http://www.omg.org/spec/BPMN/20100524/DI\" xmlns:dc=\"http://www.omg.org/spec/DD/20100524/DC\" xmlns:camunda=\"http://camunda.org/schema/1.0/bpmn\" xmlns:di=\"http://www.omg.org/spec/DD/20100524/DI\" id=\"Definitions_1\" targetNamespace=\"http://bpmn.io/schema/bpmn\"\u003e\u003cbpmn:process id=\"set_target_temp\" isExecutable=\"true\" name=\"set_target_temp\"\u003e\u003cbpmn:startEvent id=\"StartEvent_1\"\u003e\u003cbpmn:outgoing\u003eSequenceFlow_058cir1\u003c/bpmn:outgoing\u003e\u003c/bpmn:startEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_058cir1\" sourceRef=\"StartEvent_1\" targetRef=\"Task_18tgni4\"/\u003e\u003cbpmn:endEvent id=\"EndEvent_1uie5qv\"\u003e\u003cbpmn:incoming\u003eSequenceFlow_1hory47\u003c/bpmn:incoming\u003e\u003c/bpmn:endEvent\u003e\u003cbpmn:sequenceFlow id=\"SequenceFlow_1hory47\" sourceRef=\"Task_18tgni4\" targetRef=\"EndEvent_1uie5qv\"/\u003e\u003cbpmn:serviceTask id=\"Task_18tgni4\" name=\"Thermostat setTemperatureFunction\" camunda:type=\"external\" camunda:topic=\"pessimistic\"\u003e\u003cbpmn:extensionElements\u003e\u003ccamunda:inputOutput\u003e\u003ccamunda:inputParameter name=\"payload\"\u003e\u003c![CDATA[{\n\t\"function\": {\n\t\t\"id\": \"urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c\",\n\t\t\"name\": \"\",\n\t\t\"concept_id\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"characteristic_id\": \"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a\",\n\t\"device_class\": {\n\t\t\"id\": \"urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393\",\n\t\t\"name\": \"\",\n\t\t\"image\": \"\",\n\t\t\"rdf_type\": \"\"\n\t},\n\t\"device_id\": \"urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0060\",\n\t\"service_id\": \"urn:infai:ses:service:39415c76-93a3-4e8d-8740-d1a83c64bddc\",\n\t\"input\": 0\n}]]\u003e\u003c/camunda:inputParameter\u003e\u003ccamunda:inputParameter name=\"inputs\"\u003e0\u003c/camunda:inputParameter\u003e\u003c/camunda:inputOutput\u003e\u003c/bpmn:extensionElements\u003e\u003cbpmn:incoming\u003eSequenceFlow_058cir1\u003c/bpmn:incoming\u003e\u003cbpmn:outgoing\u003eSequenceFlow_1hory47\u003c/bpmn:outgoing\u003e\u003c/bpmn:serviceTask\u003e\u003c/bpmn:process\u003e\u003cbpmndi:BPMNDiagram id=\"BPMNDiagram_1\"\u003e\u003cbpmndi:BPMNPlane id=\"BPMNPlane_1\" bpmnElement=\"set_target_temp\"\u003e\u003cbpmndi:BPMNShape id=\"_BPMNShape_StartEvent_2\" bpmnElement=\"StartEvent_1\"\u003e\u003cdc:Bounds x=\"173\" y=\"102\" width=\"36\" height=\"36\"/\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_058cir1_di\" bpmnElement=\"SequenceFlow_058cir1\"\u003e\u003cdi:waypoint x=\"209\" y=\"120\"/\u003e\u003cdi:waypoint x=\"260\" y=\"120\"/\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"EndEvent_1uie5qv_di\" bpmnElement=\"EndEvent_1uie5qv\"\u003e\u003cdc:Bounds x=\"412\" y=\"102\" width=\"36\" height=\"36\"/\u003e\u003c/bpmndi:BPMNShape\u003e\u003cbpmndi:BPMNEdge id=\"SequenceFlow_1hory47_di\" bpmnElement=\"SequenceFlow_1hory47\"\u003e\u003cdi:waypoint x=\"360\" y=\"120\"/\u003e\u003cdi:waypoint x=\"412\" y=\"120\"/\u003e\u003c/bpmndi:BPMNEdge\u003e\u003cbpmndi:BPMNShape id=\"ServiceTask_1r7hcop_di\" bpmnElement=\"Task_18tgni4\"\u003e\u003cdc:Bounds x=\"260\" y=\"80\" width=\"100\" height=\"80\"/\u003e\u003c/bpmndi:BPMNShape\u003e\u003c/bpmndi:BPMNPlane\u003e\u003c/bpmndi:BPMNDiagram\u003e\u003c/bpmn:definitions\u003e","svg":"\u003c?xml version=\"1.0\" encoding=\"utf-8\"?\u003e\n\u003c!-- created with bpmn-js / http://bpmn.io --\u003e\n\u003c!DOCTYPE svg PUBLIC \"-//W3C//DTD SVG 1.1//EN\" \"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd\"\u003e\n\u003csvg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" width=\"287\" height=\"92\" viewBox=\"167 74 287 92\" version=\"1.1\"\u003e\u003cdefs\u003e\u003cmarker id=\"sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew\" viewBox=\"0 0 20 20\" refX=\"11\" refY=\"10\" markerWidth=\"10\" markerHeight=\"10\" orient=\"auto\"\u003e\u003cpath d=\"M 1 5 L 11 10 L 1 15 Z\" style=\"fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;\"/\u003e\u003c/marker\u003e\u003c/defs\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-connection\" data-element-id=\"SequenceFlow_058cir1\" style=\"display: block;\"\u003e\u003cg class=\"djs-visual\"\u003e\u003cpath d=\"m  209,120L260,120 \" style=\"fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url('#sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew');\"/\u003e\u003c/g\u003e\u003cpolyline points=\"209,120 260,120 \" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"203\" y=\"114\" width=\"63\" height=\"12\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-connection\" data-element-id=\"SequenceFlow_1hory47\" style=\"display: block;\"\u003e\u003cg class=\"djs-visual\"\u003e\u003cpath d=\"m  360,120L412,120 \" style=\"fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url('#sequenceflow-end-white-black-eydhie9pmbdsa455b0b7b3eew');\"/\u003e\u003c/g\u003e\u003cpolyline points=\"360,120 412,120 \" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"354\" y=\"114\" width=\"64\" height=\"12\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"StartEvent_1\" style=\"display: block;\" transform=\"matrix(1 0 0 1 173 102)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003ccircle cx=\"18\" cy=\"18\" r=\"18\" style=\"stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"36\" height=\"36\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"48\" height=\"48\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"EndEvent_1uie5qv\" style=\"display: block;\" transform=\"matrix(1 0 0 1 412 102)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003ccircle cx=\"18\" cy=\"18\" r=\"18\" style=\"stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"36\" height=\"36\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"48\" height=\"48\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003cg class=\"djs-group\"\u003e\u003cg class=\"djs-element djs-shape\" data-element-id=\"Task_18tgni4\" style=\"display: block;\" transform=\"matrix(1 0 0 1 260 80)\"\u003e\u003cg class=\"djs-visual\"\u003e\u003crect x=\"0\" y=\"0\" width=\"100\" height=\"80\" rx=\"10\" ry=\"10\" style=\"stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;\"/\u003e\u003ctext lineHeight=\"1.2\" class=\"djs-label\" style=\"font-family: Arial, sans-serif; font-size: 12px; font-weight: normal; fill: black;\"\u003e\u003ctspan x=\"19.3203125\" y=\"29.200000000000003\"\u003eThermostat \u003c/tspan\u003e\u003ctspan x=\"8.1484375\" y=\"43.6\"\u003esetTemperature\u003c/tspan\u003e\u003ctspan x=\"27.3203125\" y=\"58\"\u003eFunction\u003c/tspan\u003e\u003c/text\u003e\u003cpath d=\"m 12,18 v -1.71335 c 0.352326,-0.0705 0.703932,-0.17838 1.047628,-0.32133 0.344416,-0.14465 0.665822,-0.32133 0.966377,-0.52145 l 1.19431,1.18005 1.567487,-1.57688 -1.195028,-1.18014 c 0.403376,-0.61394 0.683079,-1.29908 0.825447,-2.01824 l 1.622133,-0.01 v -2.2196 l -1.636514,0.01 c -0.07333,-0.35153 -0.178319,-0.70024 -0.323564,-1.04372 -0.145244,-0.34406 -0.321407,-0.6644 -0.522735,-0.96217 l 1.131035,-1.13631 -1.583305,-1.56293 -1.129598,1.13589 c -0.614052,-0.40108 -1.302883,-0.68093 -2.022633,-0.82247 l 0.0093,-1.61852 h -2.241173 l 0.0042,1.63124 c -0.353763,0.0736 -0.705369,0.17977 -1.049785,0.32371 -0.344415,0.14437 -0.665102,0.32092 -0.9635006,0.52046 l -1.1698628,-1.15823 -1.5667691,1.5792 1.1684265,1.15669 c -0.4026573,0.61283 -0.68308,1.29797 -0.8247287,2.01713 l -1.6588041,0.003 v 2.22174 l 1.6724648,-0.006 c 0.073327,0.35077 0.1797598,0.70243 0.3242851,1.04472 0.1452428,0.34448 0.3214064,0.6644 0.5227339,0.96066 l -1.1993431,1.19723 1.5840256,1.56011 1.1964668,-1.19348 c 0.6140517,0.40346 1.3028827,0.68232 2.0233517,0.82331 l 7.19e-4,1.69892 h 2.226848 z m 0.221462,-3.9957 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 1px; stroke: black;\"/\u003e\u003cpath d=\"m 17.2,18 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 0px; stroke: black;\"/\u003e\u003cpath d=\"m 17,22 v -1.71335 c 0.352326,-0.0705 0.703932,-0.17838 1.047628,-0.32133 0.344416,-0.14465 0.665822,-0.32133 0.966377,-0.52145 l 1.19431,1.18005 1.567487,-1.57688 -1.195028,-1.18014 c 0.403376,-0.61394 0.683079,-1.29908 0.825447,-2.01824 l 1.622133,-0.01 v -2.2196 l -1.636514,0.01 c -0.07333,-0.35153 -0.178319,-0.70024 -0.323564,-1.04372 -0.145244,-0.34406 -0.321407,-0.6644 -0.522735,-0.96217 l 1.131035,-1.13631 -1.583305,-1.56293 -1.129598,1.13589 c -0.614052,-0.40108 -1.302883,-0.68093 -2.022633,-0.82247 l 0.0093,-1.61852 h -2.241173 l 0.0042,1.63124 c -0.353763,0.0736 -0.705369,0.17977 -1.049785,0.32371 -0.344415,0.14437 -0.665102,0.32092 -0.9635006,0.52046 l -1.1698628,-1.15823 -1.5667691,1.5792 1.1684265,1.15669 c -0.4026573,0.61283 -0.68308,1.29797 -0.8247287,2.01713 l -1.6588041,0.003 v 2.22174 l 1.6724648,-0.006 c 0.073327,0.35077 0.1797598,0.70243 0.3242851,1.04472 0.1452428,0.34448 0.3214064,0.6644 0.5227339,0.96066 l -1.1993431,1.19723 1.5840256,1.56011 1.1964668,-1.19348 c 0.6140517,0.40346 1.3028827,0.68232 2.0233517,0.82331 l 7.19e-4,1.69892 h 2.226848 z m 0.221462,-3.9957 c -1.788948,0.7502 -3.8576,-0.0928 -4.6097055,-1.87438 -0.7521065,-1.78321 0.090598,-3.84627 1.8802645,-4.59604 1.78823,-0.74936 3.856881,0.0929 4.608987,1.87437 0.752106,1.78165 -0.0906,3.84612 -1.879546,4.59605 z\" style=\"fill: white; stroke-width: 1px; stroke: black;\"/\u003e\u003c/g\u003e\u003crect x=\"0\" y=\"0\" width=\"100\" height=\"80\" class=\"djs-hit\" style=\"fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;\"/\u003e\u003crect x=\"-6\" y=\"-6\" width=\"112\" height=\"92\" class=\"djs-outline\" style=\"fill: none;\"/\u003e\u003c/g\u003e\u003c/g\u003e\u003c/svg\u003e"},"elements":[{"bpmn_id":"Task_18tgni4","group":null,"name":"Thermostat setTemperatureFunction","order":0,"time_event":null,"notification":null,"message_event":null,"task":{"retries":0,"parameter":{"inputs":"0"},"configurables":null,"selection":{"filter_criteria":{"characteristic_id":"urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a","function_id":"urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c","device_class_id":"urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393","aspect_id":null},"selection_options":null,"selected_device_id":"urn:infai:ses:device:dc74369e-89bc-4c7a-ad38-aa4789ea0060","selected_service_id":"urn:infai:ses:service:39415c76-93a3-4e8d-8740-d1a83c64bddc","selected_device_group_id":null}}}],"executable":false}`}}) {
		t.Error(*syncCalls)
	}
}

func strptr(s string) *string {
	return &s
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

const token = "Bearer faketoken"

func getTestPreparedDeployment(port string) (result deploymentmodel.Deployment, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		"http://localhost:"+port+"/prepared-deployments/urn:infai:ses:hub:114b6d26-5540-44e8-9aeb-234073a49995/e32329bc-3800-4429-986e-4cc208e95fc2",
		nil,
	)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != 200 {
		temp, _ := io.ReadAll(resp.Body)
		return result, errors.New(fmt.Sprint(resp.StatusCode, string(temp)))
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	result.Id = ""
	return result, nil
}

func getTestSendDeployment(port string, deployment deploymentmodel.Deployment) (err error) {
	buff := new(bytes.Buffer)
	err = json.NewEncoder(buff).Encode(deployment)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"POST",
		"http://localhost:"+port+"/deployments/urn:infai:ses:hub:114b6d26-5540-44e8-9aeb-234073a49995",
		buff,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		temp, _ := io.ReadAll(resp.Body)
		return errors.New(fmt.Sprint(resp.StatusCode, string(temp)))
	}

	return nil
}
