/*
 * Copyright 2026 InfAI (CC SES)
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

package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/SENERGY-Platform/process-deployment/lib/ctrl"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/configuration"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/controller"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/devicerepo"
	"github.com/SENERGY-Platform/process-fog-deployment/pkg/processsync"
)

// The reused ctrl starts a sync loop that calls Sync on its database; without a database
// replacement that call panics in a goroutine and ends the process.
func TestReusedCtrlSyncDoesNotPanic(t *testing.T) {
	config := &configuration.ConfigStruct{
		DeviceRepoUrl:      "http://localhost:1",
		ProcessRepoUrl:     "http://localhost:1",
		PermissionsV2Url:   "http://localhost:1",
		DeviceSelectionUrl: "http://localhost:1",
		ProcessSyncUrl:     "http://localhost:1",
	}
	c, err := controller.New(config, processsync.New(config), devicerepo.Factory)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrls := map[string]*ctrl.Ctrl{
		"ReuseCloudDeploymentWithNewDeviceRepo": c.ReuseCloudDeploymentWithNewDeviceRepo(ctx, "hub"),
		"ReuseCloudDeployment":                  c.ReuseCloudDeployment(ctx),
		"ReuseCloudDeploymentWithProcessSync":   c.ReuseCloudDeploymentWithProcessSync(ctx, "token", "hub"),
	}
	for name, reused := range ctrls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("sync panicked: %v", r)
				}
			}()
			if err, _ := reused.Sync(time.Minute).(error); err != nil {
				t.Error(err)
			}
		})
	}
}

// Each request builds a reused ctrl whose sync loop must end with the request.
func TestReusedCtrlSyncLoopEndsWithRequest(t *testing.T) {
	config := &configuration.ConfigStruct{
		DeviceRepoUrl:      "http://localhost:1",
		ProcessRepoUrl:     "http://localhost:1",
		PermissionsV2Url:   "http://localhost:1",
		DeviceSelectionUrl: "http://localhost:1",
		ProcessSyncUrl:     "http://localhost:1",
	}
	c, err := controller.New(config, processsync.New(config), devicerepo.Factory)
	if err != nil {
		t.Fatal(err)
	}

	const requests = 100
	before := runtime.NumGoroutine()
	for i := 0; i < requests; i++ {
		c.SetExecutableFlag(&deploymentmodel.Deployment{})
	}
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() >= before+requests/2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after >= before+requests/2 {
		t.Errorf("goroutines grew from %v to %v after %v requests", before, after, requests)
	}
}
