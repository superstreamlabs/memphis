// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package server

import (
	"encoding/json"
	"memphis/models"
)

const sdkClientsUpdatesSubject = "$memphis_sdk_clients_updates"

type simplifiedMsgHandler func(*client, string, string, []byte)

type memphisResponse interface {
	SetError(error)
}

type createStationRequest struct {
	StationName          string                  `json:"name"`
	SchemaName           string                  `json:"schema_name"`
	RetentionType        string                  `json:"retention_type"`
	RetentionValue       int                     `json:"retention_value"`
	StorageType          string                  `json:"storage_type"`
	Replicas             int                     `json:"replicas"`
	IdempotencyWindow    int64                   `json:"idempotency_window_in_ms"`
	DlsConfiguration     models.DlsConfiguration `json:"dls_configuration"`
	Username             string                  `json:"username"`
	TieredStorageEnabled bool                    `json:"tiered_storage_enabled"`
}

type destroyStationRequest struct {
	StationName string `json:"station_name"`
	Username    string `json:"username"`
}

type createProducerRequestV0 struct {
	Name         string `json:"name"`
	StationName  string `json:"station_name"`
	ConnectionId string `json:"connection_id"`
	ProducerType string `json:"producer_type"`
	Username     string `json:"username"`
}

type createProducerRequestV1 struct {
	Name           string `json:"name"`
	StationName    string `json:"station_name"`
	ConnectionId   string `json:"connection_id"`
	ProducerType   string `json:"producer_type"`
	RequestVersion int    `json:"req_version"`
	Username       string `json:"username"`
}

type createConsumerResponse struct {
	Err string `json:"error"`
}

type createProducerResponse struct {
	SchemaUpdate            models.ProducerSchemaUpdateInit `json:"schema_update"`
	SchemaVerseToDls        bool                            `json:"schemaverse_to_dls"`
	ClusterSendNotification bool                            `json:"send_notification"`
	Err                     string                          `json:"error"`
}

type destroyProducerRequest struct {
	StationName  string `json:"station_name"`
	ProducerName string `json:"name"`
	Username     string `json:"username"`
}

type createConsumerRequestV0 struct {
	Name             string `json:"name"`
	StationName      string `json:"station_name"`
	ConnectionId     string `json:"connection_id"`
	ConsumerType     string `json:"consumer_type"`
	ConsumerGroup    string `json:"consumers_group"`
	MaxAckTimeMillis int    `json:"max_ack_time_ms"`
	MaxMsgDeliveries int    `json:"max_msg_deliveries"`
	Username         string `json:"username"`
}

type createConsumerRequestV1 struct {
	Name                     string `json:"name"`
	StationName              string `json:"station_name"`
	ConnectionId             string `json:"connection_id"`
	ConsumerType             string `json:"consumer_type"`
	ConsumerGroup            string `json:"consumers_group"`
	MaxAckTimeMillis         int    `json:"max_ack_time_ms"`
	MaxMsgDeliveries         int    `json:"max_msg_deliveries"`
	Username                 string `json:"username"`
	StartConsumeFromSequence uint64 `json:"start_consume_from_sequence"`
	LastMessages             int64  `json:"last_messages"`
	RequestVersion           int    `json:"req_version"`
}

type attachSchemaRequest struct {
	Name        string `json:"name"`
	StationName string `json:"station_name"`
	Username    string `json:"username"`
}

type detachSchemaRequest struct {
	StationName string `json:"station_name"`
	Username    string `json:"username"`
}

type destroyConsumerRequest struct {
	StationName  string `json:"station_name"`
	ConsumerName string `json:"name"`
	Username     string `json:"username"`
}

func (cpr *createProducerResponse) SetError(err error) {
	cpr.Err = err.Error()
}

func (ccr *createConsumerResponse) SetError(err error) {
	ccr.Err = err.Error()
}

func (s *Server) initializeSDKHandlers() {
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
		func(c *client, subject, reply string, msg []byte) {
			go s.createProducerDirect(c, subject, reply, msg)
		})
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

	// schema attachements
	s.queueSubscribe("$memphis_schema_attachments",
		"memphis_schema_attachments_listeners_group",
		attachSchemaHandler(s))
	s.queueSubscribe("$memphis_schema_detachments",
		"memphis_schema_detachments_listeners_group",
		detachSchemaHandler(s))
}

func createStationHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.createStationDirect(c, reply, copyBytes(msg))
	}
}

func destroyStationHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.removeStationDirect(c, reply, copyBytes(msg))
	}
}

// func createProducerHandler() {
// return func(c *client, subject, reply string, msg []byte) {
// go s.createProducerDirect(c, reply, copyBytes(msg))
// go serv.createProducerDirect()
// go s.createProducerDirect()

// }
// }

func destroyProducerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.destroyProducerDirect(c, reply, copyBytes(msg))
	}
}

func createConsumerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.createConsumerDirect(c, reply, copyBytes(msg))
	}
}

func destroyConsumerHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.destroyConsumerDirect(c, reply, copyBytes(msg))
	}
}

func attachSchemaHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.useSchemaDirect(c, reply, copyBytes(msg))
	}
}

func detachSchemaHandler(s *Server) simplifiedMsgHandler {
	return func(c *client, subject, reply string, msg []byte) {
		go s.removeSchemaFromStationDirect(c, reply, copyBytes(msg))
	}
}

func respondWithErr(s *Server, replySubject string, err error) {
	resp := []byte("")
	if err != nil {
		resp = []byte(err.Error())
	}
	s.respondOnGlobalAcc(replySubject, resp)
}

func respondWithErrOrJsApiResp[T any](jsApi bool, c *client, acc *Account, subject, reply, msg string, resp T, err error) {
	if jsApi {
		s := c.srv
		ci := c.getClientInfo(false)
		s.sendAPIErrResponse(ci, acc, subject, reply, string(msg), s.jsonResponse(&resp))
		return
	}
	respondWithErr(c.srv, reply, err)
}

func respondWithResp(s *Server, replySubject string, resp memphisResponse) {
	rawResp, err := json.Marshal(resp)
	if err != nil {
		serv.Errorf("respondWithResp: response marshal error: " + err.Error())
		return
	}
	s.respondOnGlobalAcc(replySubject, rawResp)
}

func respondWithRespErr(s *Server, replySubject string, err error, resp memphisResponse) {
	resp.SetError(err)
	respondWithResp(s, replySubject, resp)
}

func (s *Server) SendUpdateToClients(sdkClientsUpdate models.SdkClientsUpdates) {
	subject := sdkClientsUpdatesSubject
	msg, err := json.Marshal(sdkClientsUpdate)
	if err != nil {
		s.Errorf("SendUpdateToClients: " + err.Error())
		return
	}
	s.sendInternalAccountMsg(s.GlobalAccount(), subject, msg)
}
