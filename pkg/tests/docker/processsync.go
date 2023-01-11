/*
 * Copyright 2020 InfAI (CC SES)
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

package docker

import (
	"context"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"net/http"
	"sync"
)

func ProcessSync(ctx context.Context, wg *sync.WaitGroup, permsearch string, mqttBroker string, mongoUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start process-sync")
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/process-sync",
		Tag:        "dev",
		Env: []string{
			"MONGO_URL=" + mongoUrl,
			"PERMISSIONS_URL=" + permsearch,
			"MQTT_BROKER=" + mqttBroker,
			"DEBUG=true",
			//"MQTT_GROUP_ID=sync",
			"MQTT_CLIENT_ID=sync",
		},
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{Name: "unless-stopped"}
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
	}()
	go Dockerlog(pool, ctx, container, "PROCESS-SYNC")
	hostPort = container.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try process-sync connection...")
		_, err := http.Get("http://localhost:" + hostPort)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}
