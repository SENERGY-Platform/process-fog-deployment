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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func NewStatelessRepoMock(ctx context.Context, file string) (url string, calls *map[string][]string, err error) {
	callsMap := map[string][]string{}
	calls = &callsMap
	f, err := os.Open(file)
	if err != nil {
		return url, calls, err
	}
	responses := map[string]interface{}{}
	err = json.NewDecoder(f).Decode(&responses)
	if err != nil {
		return url, calls, err
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			log.Println("ERROR: NewStatelessRepoMock()", err)
			http.Error(writer, err.Error(), 500)
			return
		}
		path := request.URL.Path
		log.Println("TEST: receive http request", file, request.Method, path)
		callsMap[path] = append(callsMap[path], strings.TrimSpace(string(payload)))
		response, ok := responses[path]
		if !ok {
			log.Println("ERROR: ", file, "unknown path", path)
			http.Error(writer, "unknown path", 500)
			return
		}
		json.NewEncoder(writer).Encode(response)
	}))
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return server.URL, calls, nil
}
