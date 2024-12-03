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

func DeviceManager(ctx context.Context, wg *sync.WaitGroup, kafkaUrl string, deviceRepoUrl string, permUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start device-manager")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/device-manager:dev",
			Env: map[string]string{
				"KAFKA_URL":          kafkaUrl,
				"PERMISSIONS_V2_URL": permUrl,
				"DEVICE_REPO_URL":    deviceRepoUrl,
				"DISABLE_VALIDATION": "true",
			},
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
		log.Println("DEBUG: remove container device-manager", c.Terminate(context.Background()))
	}()

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

func DeviceManagerWithDependencies(basectx context.Context, wg *sync.WaitGroup) (managerUrl string, repoUrl string, permUrl string, err error) {
	_, managerUrl, repoUrl, permUrl, err = DeviceManagerWithDependenciesAndKafka(basectx, wg)
	return
}

func DeviceManagerWithDependenciesAndKafka(basectx context.Context, wg *sync.WaitGroup) (kafkaUrl string, managerUrl string, repoUrl string, permUrl string, err error) {
	ctx, cancel := context.WithCancel(basectx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	_, zkIp, err := Zookeeper(ctx, wg)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}
	zookeeperUrl := zkIp + ":2181"

	kafkaUrl, err = Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}

	_, mongoIp, err := MongoDB(ctx, wg)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}

	mongoUrl := "mongodb://" + mongoIp + ":27017"

	_, permV2Ip, err := PermissionsV2(ctx, wg, mongoUrl, kafkaUrl)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}
	permUrl = "http://" + permV2Ip + ":8080"

	_, repoIp, err := DeviceRepo(ctx, wg, kafkaUrl, mongoUrl, permUrl)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}
	repoUrl = "http://" + repoIp + ":8080"

	_, managerIp, err := DeviceManager(ctx, wg, kafkaUrl, repoUrl, permUrl)
	if err != nil {
		return kafkaUrl, managerUrl, repoUrl, permUrl, err
	}
	managerUrl = "http://" + managerIp + ":8080"

	return kafkaUrl, managerUrl, repoUrl, permUrl, err
}
