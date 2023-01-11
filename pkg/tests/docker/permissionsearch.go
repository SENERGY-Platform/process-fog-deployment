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
	"time"
)

func ElasticSearch(ctx context.Context, wg *sync.WaitGroup) (hostPort string, ipAddress string, err error) {
	log.Println("start elasticsearch")
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "docker.elastic.co/elasticsearch/elasticsearch",
		Tag:        "7.6.1",
		Env: []string{
			"discovery.type=single-node",
			"path.data=/opt/elasticsearch/volatile/data",
			"path.logs=/opt/elasticsearch/volatile/logs",
		},
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{
			"/opt/elasticsearch/volatile/data": "rw",
			"/opt/elasticsearch/volatile/logs": "rw",
			"/tmp":                             "rw",
		}
	})

	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()

	hostPort = container.GetPort("9200/tcp")
	err = pool.Retry(func() error {
		log.Println("try elastic connection...")
		_, err := http.Get("http://" + container.Container.NetworkSettings.IPAddress + ":9200/_cluster/health")
		return err
	})
	if err != nil {
		log.Println(err)
	}
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}

func PermSearch(ctx context.Context, wg *sync.WaitGroup, kafkaUrl string, elasticIp string) (hostPort string, ipAddress string, err error) {
	log.Println("start permsearch")
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "ghcr.io/senergy-platform/permission-search",
		Tag:        "dev",
		Env: []string{
			"KAFKA_URL=" + kafkaUrl,
			"ELASTIC_URL=" + "http://" + elasticIp + ":9200",
			"DEBUG=true",
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
	go Dockerlog(pool, ctx, container, "PERMSEARCH")
	hostPort = container.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try permsearch connection...")
		_, err := http.Get("http://localhost:" + hostPort + "/jwt/check/devices/foo/r/bool")
		if err != nil {
			log.Println(err)
		}
		return err
	})
	if err != nil {
		return hostPort, container.Container.NetworkSettings.IPAddress, err
	}
	time.Sleep(10 * time.Second)
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}
