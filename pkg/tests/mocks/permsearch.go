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
	"encoding/json"
	"github.com/SENERGY-Platform/permission-search/lib/client"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func NewPermSearchMock(ctx context.Context) (url string, calls *map[string][]string) {
	callsMap := map[string][]string{}
	calls = &callsMap
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			log.Println("ERROR: NewPermSearchMock()", err)
			http.Error(writer, err.Error(), 500)
			return
		}
		callsMap[request.URL.Path] = append(callsMap[request.URL.Path], strings.TrimSpace(string(payload)))
		query := client.QueryMessage{}
		err = json.Unmarshal(payload, &query)
		if err != nil {
			log.Println("ERROR: NewPermSearchMock()", err)
			http.Error(writer, err.Error(), 500)
			return
		}
		result := map[string]bool{}
		if query.CheckIds == nil {
			log.Println("ERROR: NewPermSearchMock() unexpected query", query)
			http.Error(writer, "unexpected query", 500)
			return
		}
		for _, id := range query.CheckIds.Ids {
			result[id] = true
		}
		json.NewEncoder(writer).Encode(result)
	}))
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return server.URL, calls
}
