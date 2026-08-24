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

package controller

import (
	"errors"
	"time"

	"github.com/SENERGY-Platform/process-deployment/lib/model"
	"github.com/SENERGY-Platform/process-deployment/lib/model/dependencymodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/deploymentmodel"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
)

// mocks the database interface to reuse github.com/SENERGY-Platform/process-deployment/lib/ctrl
// without a mongodb. process-deployment stores a deployment and syncs it to kafka afterwards;
// for a fog deployment the write is the point where the deployment leaves for
// github.com/SENERGY-Platform/process-sync, so nothing is stored and no sync handler runs.
type DatabaseReplacement struct {
	token       string
	hubId       string
	processSync ProcessSync
}

var errDatabaseReplacement = errors.New("not supported by process-fog-deployment")

func (this *DatabaseReplacement) SetDeployment(depl messages.DeploymentCommand, syncHandler func(messages.DeploymentCommand) error) error {
	if err := validateDeployment(depl); err != nil {
		return err
	}
	return this.processSync.Deploy(this.token, this.hubId, *depl.Deployment)
}

// DeleteDeployment is only reached as a rollback of a failed SetDeployment, where nothing was stored.
func (this *DatabaseReplacement) DeleteDeployment(id string, syncDeleteHandler func(messages.DeploymentCommand) error) error {
	return nil
}

func (this *DatabaseReplacement) RetryDeploymentSync(lockduration time.Duration, syncDeleteHandler func(messages.DeploymentCommand) error, syncHandler func(messages.DeploymentCommand) error) error {
	return nil
}

func (this *DatabaseReplacement) CheckDeploymentAccess(user string, deploymentId string) (error, int) {
	return errDatabaseReplacement, 500
}

func (this *DatabaseReplacement) ListDeployments(user string, options model.DeploymentListOptions) (deployments []deploymentmodel.Deployment, err error) {
	return nil, errDatabaseReplacement
}

func (this *DatabaseReplacement) GetDeployment(user string, deploymentId string) (deployment *deploymentmodel.Deployment, err error, code int) {
	return nil, errDatabaseReplacement, 500
}

func (this *DatabaseReplacement) GetDeploymentIds(user string) (deployments []string, err error) {
	return nil, errDatabaseReplacement
}

func (this *DatabaseReplacement) GetDependencies(user string, deploymentId string) (dependencymodel.Dependencies, error, int) {
	return dependencymodel.Dependencies{}, errDatabaseReplacement, 500
}

func (this *DatabaseReplacement) GetDependenciesList(user string, limit int, offset int) ([]dependencymodel.Dependencies, error, int) {
	return nil, errDatabaseReplacement, 500
}

func (this *DatabaseReplacement) GetSelectedDependencies(user string, ids []string) ([]dependencymodel.Dependencies, error, int) {
	return nil, errDatabaseReplacement, 500
}

func (this *DatabaseReplacement) SetDependencies(dependencies dependencymodel.Dependencies) error {
	return errDatabaseReplacement
}

func (this *DatabaseReplacement) DeleteDependencies(id string) error {
	return errDatabaseReplacement
}
