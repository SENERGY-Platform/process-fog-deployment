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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"sync"
)

func DeviceRepo(ctx context.Context, wg *sync.WaitGroup, kafkaUrl string, mongoUrl string, permv2Url string) (hostPort string, ipAddress string, err error) {
	return DeviceRepoWithEnv(ctx, wg, kafkaUrl, mongoUrl, permv2Url, map[string]string{})
}

func DeviceRepoWithEnv(ctx context.Context, wg *sync.WaitGroup, kafkaUrl string, mongoUrl string, permv2Url string, env map[string]string) (hostPort string, ipAddress string, err error) {
	log.Println("start device-repository")
	env["KAFKA_URL"] = kafkaUrl
	env["PERMISSIONS_V2_URL"] = permv2Url
	env["MONGO_URL"] = mongoUrl
	//env["DEBUG"] = "true"
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "ghcr.io/senergy-platform/device-repository:dev",
			Env:             env,
			ExposedPorts:    []string{"8080/tcp"},
			WaitingFor:      wait.ForListeningPort("8080/tcp"),
			AlwaysPullImage: true,
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container device-repository", c.Terminate(context.Background()))
	}()

	/*
		err = Dockerlog(c, "DEVICE-REPOSITORY")
		if err != nil {
			return "", "", err
		}
	*/

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
