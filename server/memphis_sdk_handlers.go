// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package server

import (
	"encoding/json"
)

type simplifiedMsgHandler func(*client, string, string, []byte)

type createFactoryRequest struct {
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
}

type destroyStationRequest struct {
	StationName string `json:"station_name"`
}

type createProducerRequest struct {
	Name         string `json:"name"`
	StationName  string `json:"station_name"`
	ConnectionId string `json:"connection_id"`
	ProducerType string `json:"producer_type"`
}

type destroyProducerRequest struct {
	StationName  string `json:"station_name"`
	ProducerName string `json:"name"`
}

type createConsumerRequest struct {
	Name             string `json:"name"`
	StationName      string `json:"station_name"`
	ConnectionId     string `json:"connection_id"`
	ConsumerType     string `json:"consumer_type"`
	ConsumerGroup    string `json:"consumers_group"`
	MaxAckTimeMillis int    `json:"max_ack_time_ms"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
}

type destroyConsumerRequest struct {
	StationName  string `json:"station_name"`
	ConsumerName string `json:"name"`
}

func (s *Server) initializeSDKHandlers() {
	// factories|
	s.queueSubscribe("$memphis_factory_creations",
		"memphis_factory_creations_listeners_group",
		createFactoryHandler(s))
	s.queueSubscribe("$memphis_factory_destructions",
		"memphis_factory_destructions_listeners_group",
		destroyFactoryHandler(s))

	//stations
	s.queueSubscribe("$memphis_station_creations",
		"memphis_station_creations_listeners_group",
		createStationHandler(s))
	s.queueSubscribe("$memphis_station_destructions",
		"memphis_station_destructions_listeners_group",
		destroyStationHandler(s))

	// producers
	s.queueSubscribe("$memphis_producer_creations",
		"memphis_producer_creations_listeners_group",
		createProducerHandler(s))
	s.queueSubscribe("$memphis_producer_destructions",
		"memphis_producer_destructions_listeners_group",
		destroyProducerHandler(s))

	// consumers
	s.queueSubscribe("$memphis_consumer_creations",
		"memphis_consumer_creations_listeners_group",
		createConsumerHandler(s))
	s.queueSubscribe("$memphis_consumer_destructions",
		"memphis_consumer_destructions_listeners_group",
		destroyConsumerHandler(s))
}

func createFactoryHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var cfr createFactoryRequest
		if err := json.Unmarshal(msg, &cfr); err != nil {
			s.Warnf("failed creating factory: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}
		err := createFactoryDirect(&cfr, c)
		respondWithErr(s, reply, err)
	}
}

func destroyFactoryHandler(s *Server) simplifiedMsgHandler {
	return func(_ *client, subject, reply string, msg []byte) {
		var dfr destroyFactoryRequest
		if err := json.Unmarshal(msg, &dfr); err != nil {
			s.Warnf("failed destroying factory: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}
		err := s.RemoveFactoryDirect(&dfr)
		respondWithErr(s, reply, err)
	}
}

func createStationHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var csr createStationRequest
		if err := json.Unmarshal(msg, &csr); err != nil {
			s.Warnf("failed creating station: %v", err.Error())
			respondWithErr(s, reply, err)
			return

		}
		err := s.createStationDirect(&csr, c)
		respondWithErr(s, reply, err)
	}
}

func destroyStationHandler(s *Server) simplifiedMsgHandler {
	return func(_ *client, subject, reply string, msg []byte) {
		var dsr destroyStationRequest
		if err := json.Unmarshal(msg, &dsr); err != nil {
			s.Warnf("failed destroying station: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}

		err := s.removeStationDirect(&dsr)
		respondWithErr(s, reply, err)
	}
}

func createProducerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var cpr createProducerRequest
		if err := json.Unmarshal(msg, &cpr); err != nil {
			s.Warnf("failed creating producer: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}

		err := s.createProducerDirect(&cpr, c)
		respondWithErr(s, reply, err)
	}
}

func destroyProducerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var dpr destroyProducerRequest
		if err := json.Unmarshal(msg, &dpr); err != nil {
			s.Warnf("failed destoying producer: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}

		err := s.destroyProducerDirect(&dpr, c)
		respondWithErr(s, reply, err)
	}
}

func createConsumerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var ccr createConsumerRequest
		if err := json.Unmarshal(msg, &ccr); err != nil {
			s.Warnf("failed creating consumer: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}

		err := s.createConsumerDirect(&ccr, c)
		respondWithErr(s, reply, err)
	}
}

func destroyConsumerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		var dcr destroyConsumerRequest
		if err := json.Unmarshal(msg, &dcr); err != nil {
			s.Warnf("failed destoying consumer: %v", err.Error())
			respondWithErr(s, reply, err)
			return
		}

		err := s.destroyConsumerDirect(&dcr, c)
		respondWithErr(s, reply, err)
	}
}

func respondWithErr(s *Server, replySubject string, err error) {
	resp := []byte("")
	if err != nil {
		resp = []byte(err.Error())
	}
	s.Respond(replySubject, resp)
}
