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

type simplifiedMsgHandler func(string, string, []byte)

type createFactoryRequest struct {
	Username    string `json:"username"`
	FactoryName string `json:"factory_name"`
	FactoryDesc string `json:"factory_description"`
}

type destroyFactoryRequest struct {
	FactoryName string `json:"factory_name"`
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

type destroyStationRequest struct {
	StationName string `json:"station_name"`
}

type createProducerRequest struct {
	Name         string `json:"name"`
	StationName  string `json:"station_name"`
	ConnectionId string `json:"connection_id"`
	ProducerType string `json:"producer_type"`
	Username     string `json:"username"`
}

type destroyProducerRequest struct {
	StationName  string `json:"station_name"`
	ProducerName string `json:"name"`
	Username     string `json:"username"`
}

type createConsumerRequest struct {
	Name             string `json:"name"`
	StationName      string `json:"station_name"`
	ConnectionId     string `json:"connection_id"`
	ConsumerType     string `json:"consumer_type"`
	ConsumerGroup    string `json:"consumers_group"`
	MaxAckTimeMillis int    `json:"max_ack_time_ms"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
	Username         string `json:"username"`
}

type destroyConsumerRequest struct {
	StationName  string `json:"station_name"`
	ConsumerName string `json:"name"`
	Username     string `json:"username"`
}

func Listen(s *server.Server) {
	// factories
	s.Subscribe("$memphis_factory_creations",
		"memphis_factory_creations_subscription",
		createFactoryHandler(s))
	s.Subscribe("$memphis_factory_destructions",
		"memphis_factory_destructions_subscription",
		destroyFactoryHandler(s))

	//stations
	s.Subscribe("$memphis_station_creations",
		"memphis_station_creations_subscription",
		createStationHandler(s))
	s.Subscribe("$memphis_station_destructions",
		"memphis_station_destructions_subscription",
		destroyStationHandler(s))

	// producers
	s.Subscribe("$memphis_producer_creations",
		"memphis_producer_creations_subscription",
		createProducerHandler(s))
	s.Subscribe("$memphis_producer_destructions",
		"memphis_producer_destructions_subscription",
		destroyProducerHandler(s))

	// consumers
	s.Subscribe("$memphis_consumer_creations",
		"memphis_consumer_creations_subscription",
		createConsumerHandler(s))
	s.Subscribe("$memphis_consumer_destructions",
		"memphis_consumer_destructions_subscription",
		destroyConsumerHandler(s))
}

func createFactoryHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var cfr createFactoryRequest
		if err := json.Unmarshal(msg, &cfr); err != nil {
			s.Errorf("failed creating factory: %v", err.Error())
		}
		err := server.CreateFactoryDirect(cfr.Username, cfr.FactoryName, cfr.FactoryDesc)
		respondWithErr(s, reply, err)
	}
}

func destroyFactoryHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var dfr destroyFactoryRequest
		if err := json.Unmarshal(msg, &dfr); err != nil {
			s.Errorf("failed destroying factory: %v", err.Error())
		}
		err := s.RemoveFactoryDirect(dfr.FactoryName)
		respondWithErr(s, reply, err)
	}
}

func createStationHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var csr createStationRequest
		if err := json.Unmarshal(msg, &csr); err != nil {
			s.Errorf("failed creating station: %v", err.Error())

		}

		//TODO send csr to the func instead send all the params
		// server.CreateStationDirect(&csr)
		err := server.CreateStationDirect(s, csr.Username, csr.StationName, csr.FactoryName, csr.RetentionType,
			csr.StorageType, csr.RetentionValue, csr.Replicas, csr.DedupWindowMillis, csr.DedupEnabled)
		respondWithErr(s, reply, err)
	}
}

func destroyStationHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var dsr destroyStationRequest
		if err := json.Unmarshal(msg, &dsr); err != nil {
			s.Errorf("failed destroying station: %v", err.Error())
		}

		err := s.RemoveStationDirect(dsr.StationName)
		respondWithErr(s, reply, err)
	}
}

func createProducerHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var cpr createProducerRequest
		if err := json.Unmarshal(msg, &cpr); err != nil {
			s.Errorf("failed creating producer: %v", err.Error())
		}

		err := server.CreateProducerDirect(s, cpr.Name, cpr.StationName, cpr.ConnectionId, cpr.ProducerType, cpr.Username)
		respondWithErr(s, reply, err)
	}
}

func destroyProducerHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var dpr destroyProducerRequest
		if err := json.Unmarshal(msg, &dpr); err != nil {
			s.Errorf("failed destoying producer: %v", err.Error())
		}

		err := s.DestroyProducerDirect(dpr.StationName, dpr.ProducerName, dpr.Username)
		respondWithErr(s, reply, err)
	}
}

func createConsumerHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var ccr createConsumerRequest
		if err := json.Unmarshal(msg, &ccr); err != nil {
			s.Errorf("failed creating consumer: %v", err.Error())
		}

		err := server.CreateConsumerDirect(s, ccr.Name, ccr.StationName, ccr.ConnectionId, ccr.ConsumerType, ccr.ConsumerGroup, ccr.Username, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries)
		respondWithErr(s, reply, err)
	}
}

func destroyConsumerHandler(s *server.Server) simplifiedMsgHandler {
	return func(subject, reply string, msg []byte) {
		var dcr destroyConsumerRequest
		if err := json.Unmarshal(msg, &dcr); err != nil {
			s.Errorf("failed destoying consumer: %v", err.Error())
		}

		err := s.DestroyConsumerDirect(dcr.StationName, dcr.ConsumerName, dcr.Username)
		respondWithErr(s, reply, err)
	}
}

func respondWithErr(s *server.Server, replySubject string, err error) {
	resp := []byte("")
	if err != nil {
		resp = []byte(err.Error())
	}
	s.Respond(replySubject, resp)
}
