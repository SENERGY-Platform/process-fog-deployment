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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/process-deployment/lib/config"
	"github.com/SENERGY-Platform/process-deployment/lib/interfaces"
	"github.com/SENERGY-Platform/process-deployment/lib/model/messages"
)

// mocks sourcing interface to reuse github.com/SENERGY-Platform/process-deployment/lib/ctrl without connecting to kafka
type SourcingReplacement struct {
	token       string
	hubId       string
	processSync ProcessSync
}

func (this *SourcingReplacement) NewConsumer(ctx context.Context, config config.Config, topic string, listener func(delivery []byte) error) error {
	return nil
}

// reroutes deployment requests to github.com/SENERGY-Platform/process-sync
type ProducerReplacement struct {
	token       string
	hubId       string
	processSync ProcessSync
}

func (this *ProducerReplacement) Produce(topic string, message []byte) error {
	deplMsg := messages.DeploymentCommand{}
	err := json.Unmarshal(message, &deplMsg)
	if err != nil {
		return err
	}

	if err = validateDeployment(deplMsg); err != nil {
		return err
	}
	return this.processSync.Deploy(this.token, this.hubId, *deplMsg.Deployment)
}

func (this *SourcingReplacement) NewProducer(ctx context.Context, config config.Config, topic string) (interfaces.Producer, error) {
	return &ProducerReplacement{
		token:       this.token,
		hubId:       this.hubId,
		processSync: this.processSync,
	}, nil
}
