// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package background_tasks

import (
	"encoding/json"
	"memphis-broker/server"
)

type createFactoryRequest struct {
	Username    string `json:"username"`
	FactoryName string `json:"factory_name"`
	FactoryDesc string `json:"factory_description"`
}

type createStationRequest struct {
	StationName       string `json:"name"`
	FactoryName       string `json:"factory_name"`
	RetentionType     string `json:"retention_type"`
	RetentionValue    int    `json:"retention_value"`
	StorageType       string `json:"storage_type"`
	Replicas          int    `json:"replicas"`
	DedupEnabled      bool   `json:"dedup_enabled"`
	DedupWindowMillis int    `json:"dedup_window_in_ms"`
	Username          string `json:"username"`
}

type createProducerRequest struct {
	Name         string `json:"name"`
	StationName  string `json:"station_name"`
	ConnectionId string `json:"connection_id"`
	ProducerType string `json:"producer_type"`
	Username   	 string `json:"username"`
}

type createConsumerRequest struct {
	Name             string `json:"name"`
	StationName      string `json:"station_name"`
	ConnectionId     string `json:"connection_id"`
	ConsumerType     string `json:"consumer_type"`
	ConsumerGroup    string `json:"consumers_group"`
	MaxAckTimeMillis int    `json:"max_ack_time_ms"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
	Username    	 string `json:"username"`
}

func Listen(s *server.Server) {
	s.Subscribe("$memphis_factory_creations",
		"memphis_factory_creations_subscription",
		createFactoryHandler(s))
	s.Subscribe("$memphis_station_creations",
		"memphis_station_creations_subscription",
		createStationHandler(s))
	s.Subscribe("$memphis_producer_creations",
	"memphis_producer_creations_subscription",
	createProducerHandler(s))
	s.Subscribe("$memphis_consumer_creations",
	"memphis_consumer_creations_subscription",
	createConsumerHandler(s))
}

func createFactoryHandler(s *server.Server) func(string, []byte) {
	return func(subject string, msg []byte) {
		s.Noticef("FACTORY CREATION REQUEST!")
		var cfr createFactoryRequest
		if err := json.Unmarshal(msg, &cfr); err != nil {
			s.Errorf("failed creating factory: %v", err.Error())
		}
		server.CreateFactoryDirect(cfr.Username, cfr.FactoryName, cfr.FactoryDesc)
	}
}

func createStationHandler(s *server.Server) func(string, []byte) {
	return func(subject string, msg []byte) {
		s.Noticef("STATION CREATION REQUEST!")
		var csr createStationRequest
		if err := json.Unmarshal(msg, &csr); err != nil {
			s.Errorf("failed creating station: %v", err.Error())

		}

		server.CreateStationDirect(s, csr.Username, csr.StationName, csr.FactoryName, csr.RetentionType,
			csr.StorageType, csr.RetentionValue, csr.Replicas, csr.DedupWindowMillis, csr.DedupEnabled)

		//send csr to the func instead send all the params
		// server.CreateStationDirect(&csr)
	}
}

func createProducerHandler(s *server.Server) func(string, []byte) {
	return func(subject string, msg []byte) {
		s.Noticef("PRODUCER CREATION REQUEST!")
		var cpr createProducerRequest
		if err := json.Unmarshal(msg, &cpr); err != nil {
			s.Errorf("failed creating producer: %v", err.Error())

		}

		server.CreateProducerDirect(s, cpr.Name, cpr.StationName, cpr.ConnectionId, cpr.ProducerType, cpr.Username)
	}
}


func createConsumerHandler(s *server.Server) func(string, []byte) {
	return func(subject string, msg []byte) {
		s.Noticef("CONSUMER CREATION REQUEST!")
		var ccr createConsumerRequest
		if err := json.Unmarshal(msg, &ccr); err != nil {
			s.Errorf("failed creating producer: %v", err.Error())

		}

		server.CreateConsumerDirect(s, ccr.Name, ccr.StationName, ccr.ConnectionId, ccr.ConsumerType, ccr.ConsumerGroup, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries)
	}
}
