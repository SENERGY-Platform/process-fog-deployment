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
	"context"
	"errors"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/SENERGY-Platform/process-fog-deployment/pkg/api/util"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/SENERGY-Platform/service-commons/pkg/accesslog"
	"github.com/julienschmidt/httprouter"
)

var endpoints []func(*httprouter.Router, configuration.Config, *controller.Controller)

func Start(config configuration.Config, ctx context.Context, ctrl *controller.Controller) (err error) {
	config.GetLogger().Info("start api on " + config.ApiPort)
	router := Router(config, ctrl)
	handler := accesslog.New(util.NewCors(router))
	server := &http.Server{Addr: ":" + config.ApiPort, Handler: handler, WriteTimeout: 10 * time.Second, ReadTimeout: 2 * time.Second, ReadHeaderTimeout: 2 * time.Second}
	go func() {
		config.GetLogger().Info("listening on " + server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			config.GetLogger().Error("FATAL: api server error", "error", err)
			log.Fatal("ERROR: api server error", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err = server.Shutdown(context.Background())
		config.GetLogger().Debug("api shutdown")
	}()
	return nil
}

func Router(config configuration.Config, ctrl *controller.Controller) http.Handler {
	router := httprouter.New()
	config.GetLogger().Info("add heart beat endpoint")
	router.GET("/", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.WriteHeader(http.StatusOK)
	})
	for _, e := range endpoints {
		config.GetLogger().Info("add endpoints: " + runtime.FuncForPC(reflect.ValueOf(e).Pointer()).Name())
		e(router, config, ctrl)
	}
	return router
}
