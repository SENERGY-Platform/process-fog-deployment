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
	"testing"
	"time"

	"github.com/SENERGY-Platform/process-deployment/lib/ctrl"
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

	ctrls := map[string]*ctrl.Ctrl{
		"ReuseCloudDeploymentWithNewDeviceRepo": c.ReuseCloudDeploymentWithNewDeviceRepo("hub"),
		"ReuseCloudDeployment":                  c.ReuseCloudDeployment(),
		"ReuseCloudDeploymentWithProcessSync":   c.ReuseCloudDeploymentWithProcessSync("token", "hub"),
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
