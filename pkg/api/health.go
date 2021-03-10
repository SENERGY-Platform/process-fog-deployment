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
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, HealthEndpoints)
}

func HealthEndpoints(router *httprouter.Router, config configuration.Config, ctrl *controller.Controller) {
	router.POST("/health", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		msg, err := ioutil.ReadAll(request.Body)
		log.Println("INFO: /health", err, string(msg))
		writer.WriteHeader(http.StatusOK)
	})
}
