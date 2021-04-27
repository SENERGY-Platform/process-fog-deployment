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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel/v2"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
)

func init() {
	endpoints = append(endpoints, DeploymentEndpoints)
}

func DeploymentEndpoints(router *httprouter.Router, config configuration.Config, ctrl *controller.Controller) {
	router.POST("/prepared-deployments/:hubId", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		hubId := params.ByName("hubId")
		msg := messages.PrepareRequest{}
		err := json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			if config.Debug {
				log.Println("ERROR:", err)
			}
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, code := ctrl.PrepareDeployment(token, hubId, msg.Xml, msg.Svg)
		if err != nil {
			if config.Debug {
				log.Println("ERROR:", err)
			}
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/prepared-deployments/:hubId/:modelId", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		hubId := params.ByName("hubId")
		id := params.ByName("modelId")
		process, err, code := ctrl.GetProcessModel(token, id)
		if err != nil {
			if config.Debug {
				log.Println("ERROR:", err)
			}
			http.Error(writer, err.Error(), code)
			return
		}
		start := time.Now()
		result, err, code := ctrl.PrepareDeployment(token, hubId, process.BpmnXml, process.SvgXml)
		if err != nil {
			if config.Debug {
				log.Println("ERROR:", err)
			}
			http.Error(writer, err.Error(), code)
			return
		}
		dur := time.Now().Sub(start)
		log.Println("DEBUG: prepare deployment complete time:", dur, dur.Milliseconds())
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(writer).Encode(result)
	})

	router.POST("/deployments/:hubId", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		hubId := params.ByName("hubId")
		source := request.URL.Query().Get("source")
		deployment := deploymentmodel.Deployment{}
		err := json.NewDecoder(request.Body).Decode(&deployment)
		if err != nil {
			log.Println("ERROR: unable to parse request", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := ctrl.CreateDeployment(token, hubId, deployment, source)
		if err != nil {
			if config.Debug {
				log.Println("ERROR:", err)
			}
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(writer).Encode(true)
	})

}
