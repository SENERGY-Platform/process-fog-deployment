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

package mocks

import (
	"context"
	"github.com/SENERGY-Platform/permissions-v2/pkg/api"
	"github.com/SENERGY-Platform/permissions-v2/pkg/configuration"
	"github.com/SENERGY-Platform/permissions-v2/pkg/model"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

type PermMock struct {
	Calls *map[string][]string
}

func (this *PermMock) CheckPermission(token string, topicId string, id string, permissions ...model.Permission) (access bool, err error, code int) {
	return true, nil, http.StatusOK
}

func (this *PermMock) CheckMultiplePermissions(token string, topicId string, ids []string, permissions ...model.Permission) (access map[string]bool, err error, code int) {
	access = map[string]bool{}
	for _, id := range ids {
		access[id] = true
	}
	return access, nil, http.StatusOK
}

func (this *PermMock) ListAccessibleResourceIds(token string, topicId string, options model.ListOptions, permissions ...model.Permission) (ids []string, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) ListComputedPermissions(token string, topic string, ids []string) (result []model.ComputedPermissions, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) ListTopics(token string, options model.ListOptions) (result []model.Topic, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) GetTopic(token string, id string) (result model.Topic, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) RemoveTopic(token string, id string) (err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) SetTopic(token string, topic model.Topic) (result model.Topic, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) AdminListResourceIds(tokenStr string, topicId string, options model.ListOptions) (ids []string, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) AdminLoadFromPermissionSearch(req model.AdminLoadPermSearchRequest) (updateCount int, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) ListResourcesWithAdminPermission(token string, topicId string, options model.ListOptions) (result []model.Resource, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) GetResource(token string, topicId string, id string) (result model.Resource, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) RemoveResource(token string, topicId string, id string) (err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) SetPermission(token string, topicId string, id string, permissions model.ResourcePermissions) (result model.ResourcePermissions, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func NewPermMock(ctx context.Context) (url string, calls *map[string][]string) {
	callsMap := map[string][]string{}
	calls = &callsMap
	c := &PermMock{Calls: &callsMap}
	router := api.GetRouter(configuration.Config{}, c)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			log.Println("ERROR: NewPermMock()", err)
			http.Error(writer, err.Error(), 500)
			return
		}
		pathWithQuery := request.URL.Path + "?" + request.URL.Query().Encode()
		callsMap[pathWithQuery] = append(callsMap[pathWithQuery], strings.TrimSpace(string(payload)))
		router.ServeHTTP(writer, request)
	}))
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return server.URL, calls
}
