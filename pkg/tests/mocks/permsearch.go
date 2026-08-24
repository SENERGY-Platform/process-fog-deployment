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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/SENERGY-Platform/permissions-v2/pkg/api"
	"github.com/SENERGY-Platform/permissions-v2/pkg/configuration"
	"github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

type PermMock struct {
	Calls *map[string][]string
}

func (this *PermMock) Export(token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	//TODO implement me
	panic("implement me")
}

func (this *PermMock) Import(token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	//TODO implement me
	panic("implement me")
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

func (this *PermMock) ListTopicsContext(ctx context.Context, token string, options model.ListOptions) (result []model.Topic, err error, code int) {
	return this.ListTopics(token, options)
}

func (this *PermMock) GetTopicContext(ctx context.Context, token string, id string) (result model.Topic, err error, code int) {
	return this.GetTopic(token, id)
}

func (this *PermMock) RemoveTopicContext(ctx context.Context, token string, id string) (err error, code int) {
	return this.RemoveTopic(token, id)
}

func (this *PermMock) SetTopicContext(ctx context.Context, token string, topic model.Topic) (result model.Topic, err error, code int) {
	return this.SetTopic(token, topic)
}

func (this *PermMock) AdminListResourceIdsContext(ctx context.Context, tokenStr string, topicId string, options model.ListOptions) (ids []string, err error, code int) {
	return this.AdminListResourceIds(tokenStr, topicId, options)
}

func (this *PermMock) AdminLoadFromPermissionSearchContext(ctx context.Context, req model.AdminLoadPermSearchRequest) (updateCount int, err error, code int) {
	return this.AdminLoadFromPermissionSearch(req)
}

func (this *PermMock) ExportContext(ctx context.Context, token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	return this.Export(token, options)
}

func (this *PermMock) ImportContext(ctx context.Context, token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	return this.Import(token, importModel, options)
}

func (this *PermMock) CheckPermissionContext(ctx context.Context, token string, topicId string, id string, permissions ...model.Permission) (access bool, err error, code int) {
	return this.CheckPermission(token, topicId, id, permissions...)
}

func (this *PermMock) CheckMultiplePermissionsContext(ctx context.Context, token string, topicId string, ids []string, permissions ...model.Permission) (access map[string]bool, err error, code int) {
	return this.CheckMultiplePermissions(token, topicId, ids, permissions...)
}

func (this *PermMock) ListAccessibleResourceIdsContext(ctx context.Context, token string, topicId string, options model.ListOptions, permissions ...model.Permission) (ids []string, err error, code int) {
	return this.ListAccessibleResourceIds(token, topicId, options, permissions...)
}

func (this *PermMock) ListComputedPermissionsContext(ctx context.Context, token string, topic string, ids []string) (result []model.ComputedPermissions, err error, code int) {
	return this.ListComputedPermissions(token, topic, ids)
}

func (this *PermMock) ListResourcesWithAdminPermissionContext(ctx context.Context, token string, topicId string, options model.ListOptions) (result []model.Resource, err error, code int) {
	return this.ListResourcesWithAdminPermission(token, topicId, options)
}

func (this *PermMock) GetResourceContext(ctx context.Context, token string, topicId string, id string) (result model.Resource, err error, code int) {
	return this.GetResource(token, topicId, id)
}

func (this *PermMock) RemoveResourceContext(ctx context.Context, token string, topicId string, id string) (err error, code int) {
	return this.RemoveResource(token, topicId, id)
}

func (this *PermMock) SetPermissionContext(ctx context.Context, token string, topicId string, id string, permissions model.ResourcePermissions) (result model.ResourcePermissions, err error, code int) {
	return this.SetPermission(token, topicId, id, permissions)
}
