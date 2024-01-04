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
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/textproto"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nuid"
)

const (
	crlf      = "\r\n"
	hdrPreEnd = len(hdrLine) - len(crlf)
	statusLen = 3 // e.g. 20x, 40x, 50x
	statusHdr = "Status"
	descrHdr  = "Description"
)

const (
	syslogsStreamName           = "$memphis_syslogs"
	syslogsExternalSubject      = "extern.*"
	syslogsInfoSubject          = "extern.info"
	syslogsWarnSubject          = "extern.warn"
	syslogsErrSubject           = "extern.err"
	syslogsSysSubject           = "intern.sys"
	dlsStreamName               = "$memphis-%s-dls"
	dlsUnackedStream            = "$memphis_dls_unacked"
	dlsSchemaverseStream        = "$memphis_dls_schemaverse"
	dlsFunctionsStream          = "$memphis_dls_functions"
	dlsResendMessagesStreamNew  = "$memphis_dls_%v.%v"
	dlsResendMessagesStreamOld  = "$memphis_dls_%v_%v"
	tieredStorageStream         = "$memphis_tiered_storage"
	throughputStreamName        = "$memphis-throughput"
	throughputStreamNameV1      = "$memphis-throughput-v1"
	MEMPHIS_GLOBAL_ACCOUNT      = "$memphis"
	integrationsAuditLogsStream = "$memphis_integrations_audit_logs"
	systemTasksStreamName       = "$memphis_system_tasks"
	connectorsLogsStream        = "$memphis_connectors_logs"
	memphisSchemaDetachments    = "$memphis_schema_detachments"
	memphisConsumerCreations    = "$memphis_consumer_creations"
	memphisConsumerDestructions = "$memphis_consumer_destructions"
	memphisNotifications        = "$memphis_notifications"
	memphisProducerCreations    = "$memphis_producer_creations"
	memphisProducerDestructions = "$memphis_producer_destructions"
	memphisSchemaCreations      = "$memphis_schema_creations"
	memphisStationCreations     = "$memphis_station_creations"
	memphisStationDestructions  = "$memphis_station_destructions"
)

var noLimit = -1
var enableJetStream = true

var memphisReplaceExportString = "replaceExports"
var memphisReplaceImportString = "replaceImports"
var memphisExportString = `[
	{service: "$memphis_station_creations"},
	{service: "$memphis_station_destructions"},
	{service: "$memphis_producer_creations"},
	{service: "$memphis_producer_destructions"},
	{service: "$memphis_consumer_creations"},
	{service: "$memphis_consumer_destructions"},
	{service: "$memphis_schema_attachments"},
	{service: "$memphis_schema_detachments"},
	{service: "$memphis_schema_creations"},
	{service: "$memphis_ws_subs.>"},
	{service: "$memphis_integration_updates"},
	{service: "$memphis_notifications"},
	{service: "$memphis_schemaverse_dls"},
	{service: "$memphis_pm_acks"},
	{service: "$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>"},
	{stream: "$memphis_ws_pubs.>"},
	]
`

var memphisImportString = `[
	{service: {account: "$memphis", subject: "$memphis_station_creations"}},
	{service: {account: "$memphis", subject: "$memphis_station_destructions"}},
	{service: {account: "$memphis", subject: "$memphis_producer_creations"}},
	{service: {account: "$memphis", subject: "$memphis_producer_destructions"}},
	{service: {account: "$memphis", subject: "$memphis_consumer_creations"}},
	{service: {account: "$memphis", subject: "$memphis_consumer_destructions"}},
	{service: {account: "$memphis", subject: "$memphis_schema_attachments"}},
	{service: {account: "$memphis", subject: "$memphis_schema_detachments"}},
	{service: {account: "$memphis", subject: "$memphis_schema_creations"}},
	{service: {account: "$memphis", subject: "$memphis_ws_subs.>"}},
	{service: {account: "$memphis", subject: "$memphis_integration_updates"}},
	{service: {account: "$memphis", subject: "$memphis_notifications"}},
	{service: {account: "$memphis", subject: "$memphis_schemaverse_dls"}},
	{service: {account: "$memphis", subject: "$memphis_pm_acks"}},
	{service: {account: "$memphis", subject: "$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>"}},
	{stream: {account: "$memphis", subject: "$memphis_ws_pubs.>"}},
	]
`

// JetStream API request kinds
const (
	kindStreamInfo     = "$memphis_stream_info"
	kindCreateConsumer = "$memphis_create_consumer"
	kindDeleteConsumer = "$memphis_delete_consumer"
	kindConsumerInfo   = "$memphis_consumer_info"
	kindCreateStream   = "$memphis_create_stream"
	kindUpdateStream   = "$memphis_update_stream"
	kindDeleteStream   = "$memphis_delete_stream"
	kindDeleteMessage  = "$memphis_delete_message"
	kindPurgeStream    = "$memphis_purge_stream"
	kindStreamList     = "$memphis_stream_list"
	kindGetMsg         = "$memphis_get_msg"
	kindDeleteMsg      = "$memphis_delete_msg"
	kindPurgeAccount   = "$memphis_purge_account"
)

// errors
var (
	ErrBadHeader                           = errors.New("could not decode header")
	TIERED_STORAGE_CONSUMER_CREATED        bool
	TIERED_STORAGE_STREAM_CREATED          bool
	DLS_UNACKED_CONSUMER_CREATED           bool
	DLS_UNACKED_STREAM_CREATED             bool
	DLS_SCHEMAVERSE_STREAM_CREATED         bool
	DLS_SCHEMAVERSE_CONSUMER_CREATED       bool
	DLS_FUNCTIONS_STREAM_CREATED           bool
	DLS_FUNCTIONS_CONSUMER_CREATED         bool
	SYSLOGS_STREAM_CREATED                 bool
	THROUGHPUT_STREAM_CREATED              bool
	THROUGHPUT_LEGACY_STREAM_EXIST         bool
	INTEGRATIONS_AUDIT_LOGS_STREAM_CREATED bool
	SYSTEM_TASKS_STREAM_CREATED            bool
	FUNCTIONS_TASKS_CONSUMER_CREATED       bool
	CONNECTORS_LOGS_STREAM_CREATED         bool
)

type Messages []models.MessageDetails

func createReplyHandler(s *Server, respCh chan []byte) simplifiedMsgHandler {
	return func(_ *client, subject, _ string, msg []byte) {
		go func(msg []byte) {
			respCh <- msg
		}(copyBytes(msg))
	}
}

func jsApiRequest[R any](tenantName string, s *Server, subject, kind string, msg []byte, resp *R) error {
	// use buffered lock to limit amount of concurrent jsapi requests
	if s.memphis.jsApiMu != nil {
		s.memphis.jsApiMu.Lock()
		defer s.memphis.jsApiMu.Unlock()
	}

	account, err := s.lookupAccount(tenantName)
	if err != nil {
		return err
	}
	reply := s.getJsApiReplySubject()

	timeout := time.After(40 * time.Second)
	respCh := make(chan []byte)
	sub, err := s.subscribeOnAcc(account, reply, reply+"_sid", createReplyHandler(s, respCh))
	if err != nil {
		return err
	}

	s.sendInternalAccountMsgWithReply(account, subject, reply, nil, msg, true)

	// wait for response to arrive
	var rawResp []byte
	select {
	case rawResp = <-respCh:
		s.unsubscribeOnAcc(account, sub)
		break
	case <-timeout:
		s.unsubscribeOnAcc(account, sub)
		return fmt.Errorf("[tenant name: %v]jsapi request timeout for request type %q on %q", tenantName, kind, subject)
	}

	err = json.Unmarshal(rawResp, resp)
	if err != nil {
		s.Errorf("jsApiRequest: ", err)
	}
	return err
}

func (s *Server) getJsApiReplySubject() string {
	var sb strings.Builder
	sb.WriteString("$memphis_jsapi_reply_")
	sb.WriteString(nuid.Next())
	return sb.String()
}

func RemoveUser(username string) error {
	return nil
}

func (s *Server) WaitForLeaderElection() {
	if !s.JetStreamIsClustered() {
		return
	}

	for {
		js := s.getJetStream()
		mg := js.getMetaGroup()
		if mg == nil {
			break
		}
		ci := s.raftNodeToClusterInfo(mg)
		if ci == nil {
			break
		}

		if ci.Leader != _EMPTY_ {
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Server) CreateInternalJetStreamResources() {
	ready := !s.JetStreamIsClustered()
	retentionDur := time.Duration(s.opts.LogsRetentionDays) * time.Hour * 24

	successCh := make(chan error)

	if ready { // stand alone
		go tryCreateInternalJetStreamResources(s, retentionDur, successCh, false)
		err := <-successCh
		if err != nil {
			s.Errorf("CreateInternalJetStreamResources: system streams creation failed: %v", err.Error())
		}
	} else {
		s.WaitForLeaderElection()
		for !ready { // wait for cluster to be ready if we are in cluster mode
			timeout := time.NewTimer(1 * time.Minute)
			go tryCreateInternalJetStreamResources(s, retentionDur, successCh, true)
			select {
			case <-timeout.C:
				s.Warnf("CreateInternalJetStreamResources: system streams creation takes more than a minute")
				err := <-successCh
				if err != nil {
					s.Warnf("CreateInternalJetStreamResources: %v", err.Error())
					continue
				}
				ready = true
			case err := <-successCh:
				if err != nil {
					s.Warnf("CreateInternalJetStreamResources: %v", err.Error())
					<-timeout.C
					continue
				}
				timeout.Stop()
				ready = true
			}
		}
	}

	if s.memphis.activateSysLogsPubFunc == nil {
		s.Fatalf("internal error: sys logs publish activation func is not initialized")
	} else {
		s.memphis.activateSysLogsPubFunc()
		s.popFallbackLogs()
	}
}

func tryCreateInternalJetStreamResources(s *Server, retentionDur time.Duration, successCh chan error, isCluster bool) {
	replicas := 1
	if isCluster {
		replicas = 3
	}

	v, err := s.Varz(nil)
	if err != nil {
		successCh <- err
		return
	}

	// system logs stream
	if shouldPersistSysLogs() && !SYSLOGS_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         syslogsStreamName,
			Subjects:     []string{syslogsStreamName + ".>"},
			Retention:    LimitsPolicy,
			MaxAge:       retentionDur,
			MaxBytes:     v.JetStream.Config.MaxStore / 3, // tops third of the available storage
			MaxConsumers: -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && IsNatsErr(err, JSClusterNoPeersErrF) {
			time.Sleep(1 * time.Second)
			tryCreateInternalJetStreamResources(s, retentionDur, successCh, isCluster)
			return
		}
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		SYSLOGS_STREAM_CREATED = true
	}

	idempotencyWindow := time.Duration(1 * time.Minute)
	// tiered storage stream
	if !TIERED_STORAGE_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         tieredStorageStream,
			Subjects:     []string{tieredStorageStream + ".>"},
			Retention:    WorkQueuePolicy,
			MaxAge:       time.Hour * 24,
			MaxConsumers: -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
			Duplicates:   idempotencyWindow,
		})
		if err != nil && IsNatsErr(err, JSClusterNoPeersErrF) {
			time.Sleep(1 * time.Second)
			tryCreateInternalJetStreamResources(s, retentionDur, successCh, isCluster)
			return
		}
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		TIERED_STORAGE_STREAM_CREATED = true
	}

	// create tiered storage consumer
	if !TIERED_STORAGE_CONSUMER_CREATED {
		cc := ConsumerConfig{
			DeliverPolicy: DeliverAll,
			AckPolicy:     AckExplicit,
			Durable:       TIERED_STORAGE_CONSUMER,
			FilterSubject: tieredStorageStream + ".>",
			AckWait:       time.Duration(2) * time.Duration(s.opts.TieredStorageUploadIntervalSec) * time.Second,
			MaxAckPending: -1,
			MaxDeliver:    10,
		}
		err = serv.memphisAddConsumer(s.MemphisGlobalAccountString(), tieredStorageStream, &cc)
		if err != nil {
			successCh <- err
			return
		}
		TIERED_STORAGE_CONSUMER_CREATED = true
	}

	// dls unacked messages stream
	if !DLS_UNACKED_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         dlsUnackedStream,
			Subjects:     []string{JSAdvisoryConsumerMaxDeliveryExceedPre + ".>"},
			Retention:    WorkQueuePolicy,
			MaxAge:       time.Hour * 24,
			MaxConsumers: -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		DLS_UNACKED_STREAM_CREATED = true
	}

	// create dls unacked consumer
	if !DLS_UNACKED_CONSUMER_CREATED {
		cc := ConsumerConfig{
			DeliverPolicy: DeliverAll,
			AckPolicy:     AckExplicit,
			Durable:       DLS_UNACKED_CONSUMER,
			AckWait:       time.Duration(80) * time.Second,
			MaxAckPending: -1,
			MaxDeliver:    10,
		}
		err = serv.memphisAddConsumer(s.MemphisGlobalAccountString(), dlsUnackedStream, &cc)
		if err != nil {
			successCh <- err
			return
		}
		DLS_UNACKED_CONSUMER_CREATED = true
	}

	// create schemaverse dls stream
	if !DLS_SCHEMAVERSE_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         dlsSchemaverseStream,
			Subjects:     []string{SCHEMAVERSE_DLS_INNER_SUBJ},
			Retention:    WorkQueuePolicy,
			MaxAge:       time.Hour * 24,
			MaxConsumers: -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		DLS_SCHEMAVERSE_STREAM_CREATED = true
	}

	// create schemaverse dls consumer
	if !DLS_SCHEMAVERSE_CONSUMER_CREATED {
		cc := ConsumerConfig{
			DeliverPolicy: DeliverAll,
			AckPolicy:     AckExplicit,
			Durable:       SCHEMAVERSE_DLS_CONSUMER,
			AckWait:       time.Duration(80) * time.Second,
			MaxAckPending: -1,
			MaxDeliver:    10,
		}
		err = serv.memphisAddConsumer(s.MemphisGlobalAccountString(), dlsSchemaverseStream, &cc)
		if err != nil {
			successCh <- err
			return
		}
		DLS_SCHEMAVERSE_CONSUMER_CREATED = true
	}

	// create functions dls stream
	if shouldCreateFunctionDlsStream() && !DLS_FUNCTIONS_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         dlsFunctionsStream,
			Subjects:     []string{FUNCTIONS_DLS_INNER_SUBJ},
			Retention:    WorkQueuePolicy,
			MaxAge:       time.Hour * 24,
			MaxConsumers: -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		DLS_FUNCTIONS_STREAM_CREATED = true
	}

	// create functions dls consumer
	if shouldCreateFunctionDlsStream() && !DLS_FUNCTIONS_CONSUMER_CREATED {
		cc := ConsumerConfig{
			DeliverPolicy: DeliverAll,
			AckPolicy:     AckExplicit,
			Durable:       FUNCTIONS_DLS_CONSUMER,
			AckWait:       time.Duration(80) * time.Second,
			MaxAckPending: -1,
			MaxDeliver:    10,
		}
		err = serv.memphisAddConsumer(s.MemphisGlobalAccountString(), dlsFunctionsStream, &cc)
		if err != nil {
			successCh <- err
			return
		}
		DLS_FUNCTIONS_CONSUMER_CREATED = true
	}

	// delete the old version throughput stream
	if THROUGHPUT_LEGACY_STREAM_EXIST {
		err = s.memphisDeleteStream(s.MemphisGlobalAccountString(), throughputStreamName)
		if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
			s.Errorf("Failed deleting old internal throughput stream - %s", err.Error())
		}
	}

	// throughput kv
	if !THROUGHPUT_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         (throughputStreamNameV1),
			Subjects:     []string{throughputStreamNameV1 + ".>"},
			Retention:    LimitsPolicy,
			MaxConsumers: -1,
			MaxMsgs:      int64(-1),
			MaxBytes:     int64(-1),
			MaxAge:       time.Second * ws_updates_interval_sec, // since it stores only 1 msg per second
			Discard:      DiscardOld,
			MaxMsgsPer:   ws_updates_interval_sec,
			Storage:      FileStorage,
			Replicas:     replicas,
			NoAck:        false,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		THROUGHPUT_STREAM_CREATED = true
	}

	// create integrations audit logs stream
	if !INTEGRATIONS_AUDIT_LOGS_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         integrationsAuditLogsStream,
			Subjects:     []string{integrationsAuditLogsStream + ".>"},
			Retention:    LimitsPolicy,
			MaxAge:       time.Hour * 24 * 7, // 7 days
			MaxConsumers: -1,
			MaxMsgsPer:   200,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		INTEGRATIONS_AUDIT_LOGS_STREAM_CREATED = true
	}

	// create system tasks stream
	if shouldCreateSystemTasksStream() && !SYSTEM_TASKS_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         systemTasksStreamName,
			Subjects:     []string{systemTasksStreamName + ".>"},
			Retention:    WorkQueuePolicy,
			MaxAge:       time.Hour * 24,
			MaxConsumers: -1,
			MaxMsgsPer:   -1,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		SYSTEM_TASKS_STREAM_CREATED = true
	}

	// create function tasks consumer
	if shouldCreateSystemTasksStream() && !FUNCTIONS_TASKS_CONSUMER_CREATED {
		cc := ConsumerConfig{
			Durable:       FUNCTION_TASKS_CONSUMER,
			DeliverPolicy: DeliverAll,
			AckPolicy:     AckExplicit,
			MaxAckPending: 1,
			FilterSubject: systemTasksStreamName + ".functions",
			AckWait:       time.Duration(90) * time.Second,
			MaxDeliver:    5,
		}
		err := serv.memphisAddConsumer(serv.MemphisGlobalAccountString(), systemTasksStreamName, &cc)
		if err != nil {
			successCh <- err
			return
		}
		FUNCTIONS_TASKS_CONSUMER_CREATED = true
	}

	// create connectors logs stream
	if shouldCreateConnectorsStream() && !CONNECTORS_LOGS_STREAM_CREATED {
		err = s.memphisAddStream(s.MemphisGlobalAccountString(), &StreamConfig{
			Name:         connectorsLogsStream,
			Subjects:     []string{connectorsLogsStream + ".>"},
			Retention:    LimitsPolicy,
			MaxAge:       time.Hour * 24 * 7, // 7 days
			MaxConsumers: -1,
			MaxMsgsPer:   200,
			Discard:      DiscardOld,
			Storage:      FileStorage,
			Replicas:     replicas,
		})
		if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
			successCh <- err
			return
		}
		CONNECTORS_LOGS_STREAM_CREATED = true
	}

	successCh <- nil
}

func (s *Server) popFallbackLogs() {
	select {
	case <-s.memphis.fallbackLogQ.ch:
		break
	default:
		// if there were not fallback logs, exit
		return
	}
	logs := s.memphis.fallbackLogQ.pop()

	for _, l := range logs {
		log := l
		publishLogToSubjectAndAnalytics(s, log.label, log.log)
	}
}

func (s *Server) memphisAddStream(tenantName string, sc *StreamConfig) error {
	requestSubject := fmt.Sprintf(JSApiStreamCreateT, sc.Name)

	request, err := json.Marshal(sc)
	if err != nil {
		return err
	}

	var resp JSApiStreamCreateResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindCreateStream, request, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) memphisDeleteStream(tenantName, streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamDeleteT, streamName)

	var resp JSApiStreamCreateResponse
	err := jsApiRequest(tenantName, s, requestSubject, kindCreateStream, nil, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) memphisUpdateStream(tenantName string, sc *StreamConfig) error {
	requestSubject := fmt.Sprintf(JSApiStreamUpdateT, sc.Name)

	request, err := json.Marshal(sc)
	if err != nil {
		return err
	}

	var resp JSApiStreamUpdateResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindUpdateStream, request, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func getInternalConsumerName(cn string) string {
	return replaceDelimiters(cn)
}

func (s *Server) CreateConsumer(tenantName string, consumer models.Consumer, station models.Station, partitionsList []int) error {
	var consumerName string
	if consumer.ConsumersGroup != _EMPTY_ {
		consumerName = consumer.ConsumersGroup
	} else {
		consumerName = consumer.Name
	}

	consumerName = getInternalConsumerName(consumerName)

	var maxAckTimeMs int64
	if consumer.MaxAckTimeMs <= 0 {
		maxAckTimeMs = 30000 // 30 sec
	} else {
		maxAckTimeMs = consumer.MaxAckTimeMs
	}

	var MaxMsgDeliveries int
	if consumer.MaxMsgDeliveries <= 0 || consumer.MaxMsgDeliveries > 10 {
		MaxMsgDeliveries = 10
	} else {
		MaxMsgDeliveries = consumer.MaxMsgDeliveries
	}

	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return err
	}

	if len(partitionsList) > len(station.PartitionsList) {
		partitionsList = station.PartitionsList
	}

	var deliveryPolicy DeliverPolicy
	var optStartSeq uint64
	// This check for case when the last message is 0 (in case StartConsumeFromSequence > 1 the LastMessages is 0 )
	if consumer.LastMessages == 0 && consumer.StartConsumeFromSeq == 0 {
		deliveryPolicy = DeliverNew
	} else if consumer.LastMessages > 0 {
		streamInfo, err := serv.memphisStreamInfo(tenantName, stationName.Intern())
		if err != nil {
			return err
		}
		lastSeq := streamInfo.State.LastSeq
		lastMessages := (lastSeq - uint64(consumer.LastMessages)) + 1
		if int(lastMessages) < 1 {
			lastMessages = uint64(1)
		}
		deliveryPolicy = DeliverByStartSequence
		optStartSeq = lastMessages
	} else if consumer.StartConsumeFromSeq > 1 {
		deliveryPolicy = DeliverByStartSequence
		optStartSeq = consumer.StartConsumeFromSeq
	} else if consumer.StartConsumeFromSeq == 1 || consumer.LastMessages == -1 {
		deliveryPolicy = DeliverAll
	}
	if len(partitionsList) == 0 {
		consumerConfig := &ConsumerConfig{
			Durable:       consumerName,
			DeliverPolicy: deliveryPolicy,
			AckPolicy:     AckExplicit,
			AckWait:       time.Duration(maxAckTimeMs) * time.Millisecond,
			MaxDeliver:    MaxMsgDeliveries,
			FilterSubject: stationName.Intern() + ".final",
			ReplayPolicy:  ReplayInstant,
			MaxAckPending: -1,
			HeadersOnly:   false,
			// RateLimit: ,// Bits per sec
			// Heartbeat: // time.Duration,
		}

		if deliveryPolicy == DeliverByStartSequence {
			consumerConfig.OptStartSeq = optStartSeq
		}
		err = s.memphisAddConsumer(tenantName, stationName.Intern(), consumerConfig)
		return err
	} else {
		for _, pl := range partitionsList {
			consumerConfig := &ConsumerConfig{
				Durable:       consumerName,
				DeliverPolicy: deliveryPolicy,
				AckPolicy:     AckExplicit,
				AckWait:       time.Duration(maxAckTimeMs) * time.Millisecond,
				MaxDeliver:    MaxMsgDeliveries,
				FilterSubject: stationName.Intern() + "$" + strconv.Itoa(pl) + ".final",
				ReplayPolicy:  ReplayInstant,
				MaxAckPending: -1,
				HeadersOnly:   false,
				// RateLimit: ,// Bits per sec
				// Heartbeat: // time.Duration,
			}

			if deliveryPolicy == DeliverByStartSequence {
				consumerConfig.OptStartSeq = optStartSeq
			}
			err = s.memphisAddConsumer(tenantName, stationName.Intern()+"$"+strconv.Itoa(pl), consumerConfig)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) memphisAddConsumer(tenantName, streamName string, cc *ConsumerConfig) error {
	requestSubject := fmt.Sprintf(JSApiConsumerCreateT, streamName)
	if cc.Durable != _EMPTY_ {
		requestSubject = fmt.Sprintf(JSApiDurableCreateT, streamName, cc.Durable)
	}

	request := CreateConsumerRequest{Stream: streamName, Config: *cc}
	rawRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}
	var resp JSApiConsumerCreateResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindCreateConsumer, []byte(rawRequest), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) RemoveConsumer(tenantName string, stationName StationName, cn string, partitionsList []int) error {
	cn = getInternalConsumerName(cn)
	if len(partitionsList) == 0 {
		return s.memphisRemoveConsumer(tenantName, stationName.Intern(), cn)
	} else {
		for _, pl := range partitionsList {
			err := s.memphisRemoveConsumer(tenantName, stationName.Intern()+"$"+strconv.Itoa(pl), cn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) memphisRemoveConsumer(tenantName, streamName, cn string) error {
	requestSubject := fmt.Sprintf(JSApiConsumerDeleteT, streamName, cn)
	var resp JSApiConsumerDeleteResponse
	err := jsApiRequest(tenantName, s, requestSubject, kindDeleteConsumer, []byte(_EMPTY_), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) GetCgInfo(tenantName string, stationName StationName, cgName string, partitionsList []int) (ConsumerInfo, error) {
	var resp JSApiConsumerInfoResponse
	cgName = replaceDelimiters(cgName)
	var cgInfo ConsumerInfo
	if len(partitionsList) == 0 {
		requestSubject := fmt.Sprintf(JSApiConsumerInfoT, stationName.Intern(), cgName)
		err := jsApiRequest(tenantName, s, requestSubject, kindConsumerInfo, []byte(_EMPTY_), &resp)
		if err != nil {
			return ConsumerInfo{}, err
		}
		err = resp.ToError()
		if err != nil {
			return ConsumerInfo{}, err
		}
		cgInfo = *resp.ConsumerInfo
	} else {
		init := false
		for _, pl := range partitionsList {
			stationWithPartition := fmt.Sprintf("%s$%s", stationName.Intern(), strconv.Itoa(pl))
			requestSubject := fmt.Sprintf(JSApiConsumerInfoT, stationWithPartition, cgName)
			err := jsApiRequest(tenantName, s, requestSubject, kindConsumerInfo, []byte(_EMPTY_), &resp)
			if err != nil {
				return ConsumerInfo{}, err
			}
			err = resp.ToError()
			if err != nil {
				return ConsumerInfo{}, err
			}
			if !init {
				cgInfo = *resp.ConsumerInfo
				init = true
			} else {
				cgInfo.NumAckPending += resp.ConsumerInfo.NumAckPending
				cgInfo.NumRedelivered += resp.ConsumerInfo.NumRedelivered
				cgInfo.NumWaiting += resp.ConsumerInfo.NumWaiting
				cgInfo.NumPending += resp.ConsumerInfo.NumPending
			}
		}
	}

	return cgInfo, nil
}

func (s *Server) RemoveStream(tenantName, streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamDeleteT, streamName)

	var resp JSApiStreamDeleteResponse
	err := jsApiRequest(tenantName, s, requestSubject, kindDeleteStream, []byte(_EMPTY_), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) PurgeStream(tenantName, streamName string, partitionNumber int) error {
	var streamAndPartition string
	if partitionNumber == -1 {
		streamAndPartition = streamName
	} else {
		streamAndPartition = fmt.Sprintf("%v$%v", streamName, partitionNumber)
	}
	requestSubject := fmt.Sprintf(JSApiStreamPurgeT, streamAndPartition)

	req := JSApiStreamPurgeRequest{Subject: streamAndPartition + ".final"}
	rawRequest, err := json.Marshal(req)
	if err != nil {
		return err
	}
	var resp JSApiStreamPurgeResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindPurgeStream, rawRequest, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) Opts() *Options {
	return s.opts
}

func (s *Server) MemphisVersion() string {
	data, _ := os.ReadFile("version.conf")
	return string(data)
}

func (s *Server) RemoveMsg(tenantName string, stationName StationName, msgSeq uint64, partitionNumber int) error {
	var streamName string
	if partitionNumber == -1 {
		streamName = stationName.Intern()
	} else {
		streamName = fmt.Sprintf("%v$%v", stationName.Intern(), partitionNumber)
	}
	requestSubject := fmt.Sprintf(JSApiMsgDeleteT, streamName)

	var resp JSApiMsgDeleteResponse
	req := JSApiMsgDeleteRequest{Seq: msgSeq}
	reqj, _ := json.Marshal(req)
	err := jsApiRequest(tenantName, s, requestSubject, kindDeleteMessage, reqj, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) GetTotalMessagesInStation(tenantName string, streamName string) (int, error) {
	streamInfo, err := s.memphisStreamInfo(tenantName, streamName)
	if err != nil {
		return 0, err
	}

	return int(streamInfo.State.Subjects[streamName+".final"]), nil
}

// low level call, call only with internal station name (i.e stream name)!
func (s *Server) memphisStreamInfo(tenantName string, streamName string) (*StreamInfo, error) {
	requestSubject := fmt.Sprintf(JSApiStreamInfoT, streamName)

	var resp JSApiStreamInfoResponse
	req := JSApiStreamInfoRequest{SubjectsFilter: streamName + ".>"}
	rawRequest, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = jsApiRequest(tenantName, s, requestSubject, kindStreamInfo, rawRequest, &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.StreamInfo, nil
}

func (s *Server) memphisPurgeResourcesAccount(tenantName string) error {
	requestSubject := fmt.Sprintf(JSApiAccountPurgeT, tenantName)

	var resp JSApiAccountPurgeResponse
	err := jsApiRequest(tenantName, s, requestSubject, kindPurgeAccount, nil, &resp)
	if err != nil {
		return err
	}

	err = resp.ToError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GetAvgMsgSizeInStation(station models.Station) (int64, error) {
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return 0, err
	}

	var msgBytes uint64
	var msgCount uint64
	if len(station.PartitionsList) == 0 {
		streamInfo, err := s.memphisStreamInfo(station.TenantName, stationName.Intern())
		if err != nil || streamInfo.State.Bytes == 0 {
			return 0, err
		}
		msgBytes = streamInfo.State.Bytes
		msgCount = streamInfo.State.Msgs
	} else {
		for _, p := range station.PartitionsList {
			streamInfo, err := s.memphisStreamInfo(station.TenantName, fmt.Sprintf("%v$%v", stationName.Intern(), p))
			if err != nil {
				return 0, err
			}
			msgBytes = msgBytes + streamInfo.State.Bytes
			msgCount = msgCount + streamInfo.State.Msgs
		}
	}

	if err != nil || msgBytes == 0 {
		return 0, err
	}

	return int64(msgBytes / msgCount), nil
}

func (s *Server) GetAvgMsgSizeInPartition(tenantName, streamName string) (int64, error) {
	var msgBytes uint64
	var msgCount uint64

	streamInfo, err := s.memphisStreamInfo(tenantName, streamName)
	if err != nil || streamInfo.State.Bytes == 0 {
		return 0, err
	}
	msgBytes = streamInfo.State.Bytes
	msgCount = streamInfo.State.Msgs

	return int64(msgBytes / msgCount), nil
}

func (s *Server) memphisAllStreamsInfo(tenantName string) ([]*StreamInfo, error) {
	requestSubject := JSApiStreamList
	streams := make([]*StreamInfo, 0)

	offset := 0
	offsetReq := ApiPagedRequest{Offset: offset}
	request := JSApiStreamListRequest{ApiPagedRequest: offsetReq, Subject: "*.>"}
	rawRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var resp JSApiStreamListResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindStreamList, []byte(rawRequest), &resp)
	if err != nil {
		return nil, err
	}
	err = resp.ToError()
	if err != nil {
		return nil, err
	}
	streams = append(streams, resp.Streams...)

	for len(streams) < resp.Total {
		offset += resp.Limit
		offsetReq := ApiPagedRequest{Offset: offset}
		request := JSApiStreamListRequest{ApiPagedRequest: offsetReq}
		rawRequest, err := json.Marshal(request)
		if err != nil {
			return nil, err
		}

		err = jsApiRequest(tenantName, s, requestSubject, kindStreamList, []byte(rawRequest), &resp)
		if err != nil {
			return nil, err
		}
		err = resp.ToError()
		if err != nil {
			return nil, err
		}

		streams = append(streams, resp.Streams...)
	}

	return streams, nil
}

func (s *Server) GetMessages(station models.Station, messagesToFetch int) ([]models.MessageDetails, error) {
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return []models.MessageDetails{}, err
	}

	if len(station.PartitionsList) == 0 {
		return s.GetMessagesFromPartition(station, stationName.Intern(), messagesToFetch, 0)
	} else {
		var messages Messages
		for _, p := range station.PartitionsList {
			partitionMessages, err := s.GetMessagesFromPartition(station, fmt.Sprintf("%v$%v", stationName.Intern(), p), messagesToFetch, p)
			if err != nil {
				return []models.MessageDetails{}, err
			}
			messages = append(messages, partitionMessages...)
		}

		if len(messages) == 0 {
			return []models.MessageDetails{}, nil
		} else if len(messages) <= 1000 {
			return messages, nil
		} else {
			sort.Sort(messages)
			return messages[:1000], nil
		}
	}
}

func (msgs Messages) Len() int {
	return len(msgs)
}

func (msgs Messages) Less(i, j int) bool {
	return msgs[i].TimeSent.Before(msgs[j].TimeSent)
}

func (msgs Messages) Swap(i, j int) {
	msgs[i], msgs[j] = msgs[j], msgs[i]
}

func (s *Server) GetMessagesFromPartition(station models.Station, streamName string, messagesToFetch int, partition int) ([]models.MessageDetails, error) {
	streamInfo, err := s.memphisStreamInfo(station.TenantName, streamName)
	if err != nil {
		return []models.MessageDetails{}, err
	}
	totalMessages := streamInfo.State.Subjects[streamName+".final"]
	lastStreamSeq := streamInfo.State.SubjectsState[streamName+".final"].Last

	var startSequence uint64 = 1
	if totalMessages > uint64(messagesToFetch) {
		startSequence = lastStreamSeq - uint64(messagesToFetch) + 1
	} else {
		messagesToFetch = int(totalMessages)
	}

	filterSubj := streamName + ".final"
	if !station.IsNative {
		filterSubj = _EMPTY_
	}
	replicas := 1
	if streamInfo.Config.Retention == InterestPolicy {
		replicas = streamInfo.Config.Replicas
	}
	msgs, err := s.memphisGetMsgs(station.TenantName, filterSubj,
		streamName,
		startSequence,
		messagesToFetch,
		5*time.Second,
		true,
		station.RetentionType == "ack_based",
		replicas,
	)

	if err != nil {
		return []models.MessageDetails{}, err
	}

	if len(msgs) == 0 {
		return []models.MessageDetails{}, nil
	}

	var messages []models.MessageDetails

	stationIsNative := station.IsNative

	for _, msg := range msgs {
		messageDetails := models.MessageDetails{
			MessageSeq: int(msg.Sequence),
			TimeSent:   msg.Time,
			Size:       len(msg.Subject) + len(msg.Data) + len(msg.Header),
		}

		data := hex.EncodeToString(msg.Data)
		if len(data) > 80 { // get the first chars for preview needs
			data = data[0:80]
		}
		messageDetails.Data = data

		var headersJson map[string]string
		if stationIsNative {
			if msg.Header != nil {
				headersJson, err = DecodeHeader(msg.Header)
				if err != nil {
					return nil, err
				}
			}
			connectionIdHeader := headersJson["$memphis_connectionId"]
			producedByHeader := strings.ToLower(headersJson["$memphis_producedBy"])

			for header := range headersJson {
				if strings.HasPrefix(header, MEMPHIS_GLOBAL_ACCOUNT) {
					delete(headersJson, header)
				}
			}

			if producedByHeader == "$memphis_dls" { // skip poison messages which have been resent
				continue
			}
			messageDetails.ProducedBy = producedByHeader
			messageDetails.ConnectionId = connectionIdHeader
			messageDetails.Headers = headersJson
			messageDetails.Partition = partition
		}

		messages = append(messages, messageDetails)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].MessageSeq < messages[j].MessageSeq
	})

	return messages, nil
}

func getHdrLastIdxFromRaw(msg []byte) int {
	inCrlf := false
	inDouble := false
	for i, b := range msg {
		switch b {
		case '\r':
			inCrlf = true
		case '\n':
			if inDouble {
				return i
			}
			inDouble = inCrlf
			inCrlf = false
		default:
			inCrlf, inDouble = false, false
		}
	}
	return -1
}

func (s *Server) memphisGetMsgs(tenantName, filterSubj, streamName string, startSeq uint64, amount int, timeout time.Duration, findHeader, isAckBasedStation bool, consumerReplicas int) ([]StoredMsg, error) {
	uid, _ := uuid.NewV4()
	durableName := "$memphis_fetch_messages_consumer_" + uid.String()

	cc := ConsumerConfig{
		FilterSubject: filterSubj,
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		Name:          durableName,
		AckPolicy:     AckNone,
		Replicas:      consumerReplicas,
	}

	err := s.memphisAddConsumer(tenantName, streamName, &cc)
	if err != nil {
		return nil, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, streamName, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.Itoa(amount))

	account, err := s.lookupAccount(tenantName)
	if err != nil {
		return nil, err
	}

	sub, err := s.subscribeOnAcc(account, reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, reply string, msg []byte, findHeader bool) {
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf("memphisGetMsgs: %v", err.Error())
				return
			}

			dataFirstIdx := 0
			dataLen := len(msg)
			if findHeader {
				dataFirstIdx = getHdrLastIdxFromRaw(msg) + 1
				if dataFirstIdx > len(msg)-len(CR_LF) {
					s.Errorf("memphisGetMsgs: memphis error parsing in station get messages")
					return
				}

				dataLen = len(msg) - dataFirstIdx
			}
			dataLen -= len(CR_LF)

			respCh <- StoredMsg{
				Sequence: uint64(seq),
				Header:   msg[:dataFirstIdx],
				Data:     msg[dataFirstIdx : dataFirstIdx+dataLen],
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, reply, copyBytes(msg), findHeader)
	})
	if err != nil {
		return nil, err
	}

	s.sendInternalAccountMsgWithReply(account, subject, reply, nil, req, true)

	var msgs []StoredMsg
	timer := time.NewTimer(timeout)
	for i := 0; i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	s.unsubscribeOnAcc(account, sub)

	return msgs, nil
}

func (s *Server) GetMessage(tenantName string, stationName StationName, msgSeq uint64, paritionNumber int) (*StoredMsg, error) {
	var streamName string
	if paritionNumber != -1 {
		streamName = fmt.Sprintf("%v$%v", stationName.Intern(), paritionNumber)
	} else {
		streamName = stationName.Intern()
	}

	return s.memphisGetMessage(tenantName, streamName, msgSeq)
}

func (s *Server) GetLeaderAndFollowers(station models.Station, partitionNumber int) (string, []string, error) {
	var followers []string
	var streamInfo *StreamInfo
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return _EMPTY_, []string{}, err
	}
	if len(station.PartitionsList) > 0 {
		if partitionNumber == -1 {
			return _EMPTY_, []string{}, nil
		}
		streamName := fmt.Sprintf("%s$%s", stationName.Intern(), strconv.Itoa(partitionNumber))
		streamInfo, err = s.memphisStreamInfo(station.TenantName, streamName)
		if err != nil {
			return _EMPTY_, []string{}, err
		}
	} else { // backward compatibility
		streamInfo, err = s.memphisStreamInfo(station.TenantName, stationName.Intern())
		if err != nil {
			return _EMPTY_, []string{}, err
		}

	}
	for _, replica := range streamInfo.Cluster.Replicas {
		followers = append(followers, replica.Name)
	}

	return streamInfo.Cluster.Leader, followers, nil
}

func (s *Server) memphisGetMessage(tenantName, streamName string, msgSeq uint64) (*StoredMsg, error) {
	requestSubject := fmt.Sprintf(JSApiMsgGetT, streamName)
	request := JSApiMsgGetRequest{Seq: msgSeq}
	rawRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var resp JSApiMsgGetResponse
	err = jsApiRequest(tenantName, s, requestSubject, kindGetMsg, rawRequest, &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.Message, nil
}

func (s *Server) queueSubscribe(tenantName string, subj, queueGroupName string, cb simplifiedMsgHandler) error {
	acc, err := s.lookupAccount(tenantName)
	if err != nil {
		return err
	}

	acc.mu.Lock()
	c := acc.internalClient()

	acc.isid++
	sid := strconv.FormatUint(acc.isid, 10)
	acc.mu.Unlock()

	wcb := func(_ *subscription, c *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(c, subject, reply, rmsg)
	}

	_, err = c.processSub([]byte(subj), []byte(queueGroupName), []byte(sid), wcb, false)

	return err
}

func (s *Server) subscribeOnAcc(acc *Account, subj, sid string, cb simplifiedMsgHandler) (*subscription, error) {
	acc.mu.Lock()
	c := acc.internalClient()
	acc.mu.Unlock()

	wcb := func(_ *subscription, c *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(c, subject, reply, rmsg)
	}

	return c.processSub([]byte(subj), nil, []byte(sid), wcb, false)
}

func (s *Server) unsubscribeOnAcc(acc *Account, sub *subscription) error {
	acc.mu.Lock()
	c := acc.internalClient()
	acc.mu.Unlock()
	return c.processUnsub(sub.sid)
}

func (s *Server) ResendPoisonMessage(tenantName, subject string, data, headers []byte) error {
	hdrs := make(map[string]string)
	err := json.Unmarshal(headers, &hdrs)
	if err != nil {
		return err
	}

	hdrs["$memphis_producedBy"] = "$memphis_dls"

	if hdrs["producedBy"] != _EMPTY_ {
		delete(hdrs, "producedBy")
	}

	account, err := s.lookupAccount(tenantName)
	if err != nil {
		return err
	}

	s.sendInternalMsgWithHeaderLocked(account, subject, hdrs, data)
	return nil
}

func (s *Server) sendInternalMsgWithHeaderLocked(acc *Account, subj string, hdr map[string]string, msg interface{}) {

	acc.mu.Lock()
	c := acc.internalClient()
	acc.mu.Unlock()

	s.mu.Lock()
	if s.sys == nil || s.sys.sendq == nil {
		return
	}
	s.sys.sendq.push(newPubMsg(c, subj, _EMPTY_, nil, hdr, msg, noCompression, false, false))
	s.mu.Unlock()
}

func DecodeHeader(buf []byte) (map[string]string, error) {
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(buf)))
	l, err := tp.ReadLine()
	hdr := make(map[string]string)
	if l == _EMPTY_ {
		return hdr, nil
	}

	if err != nil || len(l) < hdrPreEnd || l[:hdrPreEnd] != hdrLine[:hdrPreEnd] {
		return nil, ErrBadHeader
	}

	// tp.readMIMEHeader changes key cases
	mh, err := readMIMEHeader(tp)
	if err != nil {
		return nil, err
	}

	// Check if we have an inlined status.
	if len(l) > hdrPreEnd {
		var description string
		status := strings.TrimSpace(l[hdrPreEnd:])
		if len(status) != statusLen {
			description = strings.TrimSpace(status[statusLen:])
			status = status[:statusLen]
		}
		mh.Add(statusHdr, status)
		if len(description) > 0 {
			mh.Add(descrHdr, description)
		}
	}

	for k, v := range mh {
		hdr[k] = v[0]
	}
	return hdr, nil
}

// readMIMEHeader returns a MIMEHeader that preserves the
// original case of the MIME header, based on the implementation
// of textproto.ReadMIMEHeader.
//
// https://golang.org/pkg/net/textproto/#Reader.ReadMIMEHeader
func readMIMEHeader(tp *textproto.Reader) (textproto.MIMEHeader, error) {
	m := make(textproto.MIMEHeader)
	for {
		kv, err := tp.ReadLine()
		if len(kv) == 0 {
			return m, err
		}

		// Process key fetching original case.
		i := strings.IndexByte(kv, ':')
		if i < 0 {
			return nil, ErrBadHeader
		}
		key := kv[:i]
		if key == _EMPTY_ {
			// Skip empty keys.
			continue
		}
		i++
		for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
			i++
		}
		value := string(kv[i:])
		m[key] = append(m[key], value)
		if err != nil {
			return m, err
		}
	}
}

func GetMemphisOpts(opts *Options) (*Options, error) {
	_, configs, err := db.GetAllConfigurations()
	if err != nil {
		return nil, err
	}

	if opts.DlsRetentionHours == nil { // for cases the broker started without config file
		opts.DlsRetentionHours = make(map[string]int)
	}

	if opts.GCProducersConsumersRetentionHours == nil { // for cases the broker started without config file
		opts.GCProducersConsumersRetentionHours = make(map[string]int)
	}

	for _, conf := range configs {
		switch conf.Key {
		case "dls_retention":
			v, _ := strconv.Atoi(conf.Value)
			opts.DlsRetentionHours[conf.TenantName] = v
		case "logs_retention":
			v, _ := strconv.Atoi(conf.Value)
			opts.LogsRetentionDays = v
		case "tiered_storage_time_sec":
			v, _ := strconv.Atoi(conf.Value)
			opts.TieredStorageUploadIntervalSec = v
		case "ui_host":
			opts.UiHost = conf.Value
		case "broker_host":
			opts.BrokerHost = conf.Value
		case "rest_gw_host":
			opts.RestGwHost = conf.Value
		case "max_msg_size_mb":
			v, _ := strconv.Atoi(conf.Value)
			opts.MaxPayload = int32(v * 1024 * 1024)
		case "gc_producer_consumer_retention_hours":
			v, _ := strconv.Atoi(conf.Value)
			opts.GCProducersConsumersRetentionHours[conf.TenantName] = v
		}
	}

	return opts, nil
}

func (s *Server) getTenantNameAndMessage(msg []byte) (string, string, error) {
	var ci ClientInfo
	var tenantName string
	message := string(msg)

	hdr := getHeader(ClientInfoHdr, msg)
	if len(hdr) > 0 {
		if err := json.Unmarshal(hdr, &ci); err != nil {
			s.Errorf("getTenantNameAndMessage: ", err)
			return tenantName, message, err
		}
		tenantName = ci.Account
		message = message[len(hdrLine)+len(ClientInfoHdr)+len(hdr)+6:]
	} else {
		tenantName = s.MemphisGlobalAccountString()
	}

	return tenantName, message, nil
}

func generateRandomPassword(length int) string {
	allowedPasswordChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$"
	charsetLength := big.NewInt(int64(len(allowedPasswordChars)))
	password := make([]byte, length)

	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, charsetLength)
		password[i] = allowedPasswordChars[randomIndex.Int64()]
	}

	return string(password)
}

type UserConfig struct {
	User        string          `json:"user"`
	Password    string          `json:"password"`
	Permissions NatsPermissions `json:"permissions,omitempty"`
}

type NatsPermissions struct {
	Publish   NatsAuthorization `json:"publish,omitempty"`
	Subscribe NatsAuthorization `json:"subscribe,omitempty"`
}

type NatsAuthorization struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type AccountConfig struct {
	Jetstream *bool           `json:"jetstream,omitempty"`
	Users     []UserConfig    `json:"users,omitempty"`
	Limits    map[string]*int `json:"limits,omitempty"`
	Exports   string          `json:"exports,omitempty"`
	Imports   string          `json:"imports,omitempty"`
}

type Authorization struct {
	Users []UserConfig `json:"users,omitempty"`
}

type Data struct {
	Accounts map[string]AccountConfig `json:"accounts,omitempty"`
}

func generateJSONString(accounts map[string]AccountConfig) (string, error) {
	data := Data{
		Accounts: accounts,
	}

	jsonString, err := json.MarshalIndent(data, " ", _EMPTY_)
	if err != nil {
		return _EMPTY_, err
	}
	var dataMap map[string]interface{}
	err = json.Unmarshal(jsonString, &dataMap)
	if err != nil {
		panic(err)
	}
	newStr := string(jsonString)[1 : len(string(jsonString))-1]
	return newStr, nil
}

func getAccountsAndUsersString() (string, error) {
	decriptionKey := getAESKey()
	users, err := db.GetAllUsersByType([]string{"application"})
	if err != nil {
		return _EMPTY_, err
	}

	tenants, err := db.GetAllTenantsWithoutGlobal()
	if err != nil {
		return _EMPTY_, err
	}
	globalUsers := []UserConfig{{User: "$$memphis", Password: getInternalUserPassword()}}
	accounts := map[string]AccountConfig{
		"$SYS": {
			Users: []UserConfig{
				{User: "$SYS", Password: getInternalUserPassword()},
			},
			Limits: map[string]*int{"max_connections": &noLimit}},
	}
	if shouldCreateRootUserforGlobalAcc {
		_, globalT, err := db.GetGlobalTenant()
		if err != nil {
			return _EMPTY_, err
		}
		decryptedPass, err := DecryptAES(decriptionKey, globalT.InternalWSPass)
		if err != nil {
			return _EMPTY_, err
		}
		globalUsers = append(globalUsers, UserConfig{User: "$memphis_user$1", Password: decryptedPass})
		globalUsers = append(globalUsers, UserConfig{User: "root$1", Password: configuration.ROOT_PASSWORD})
	}
	tenantsToUsers := map[string][]UserConfig{}
	for _, user := range users {
		tName := user.TenantName
		decryptedUserPassword, err := DecryptAES(decriptionKey, user.Password)
		if err != nil {
			return _EMPTY_, err
		}
		if tName == MEMPHIS_GLOBAL_ACCOUNT {
			if len(user.Roles) > 0 {
				allowReadSubjects, allowWriteSubjects, err := GetAllowedSubjectsFromRoleIds(user.Roles, tName)
				if err != nil {
					//return _EMPTY_, err
					fmt.Printf("user: %v, err: %v\n", user.Username, err)
				}

				globalUsers = append(globalUsers, UserConfig{
					User:     user.Username + "$1",
					Password: decryptedUserPassword,
					Permissions: NatsPermissions{
						Publish:   NatsAuthorization{Allow: allowWriteSubjects},
						Subscribe: NatsAuthorization{Allow: allowReadSubjects},
					},
				})
			} else {
				globalUsers = append(globalUsers, UserConfig{User: user.Username + "$1", Password: decryptedUserPassword})
			}
			continue
		}
		if len(user.Roles) > 0 {
			allowReadSubjects, allowWriteSubjects, err := GetAllowedSubjectsFromRoleIds(user.Roles, tName)
			if err != nil {
				//return _EMPTY_, err
				fmt.Printf("user: %v, err: %v\n", user.Username, err)
			}

			if usrMap, ok := tenantsToUsers[tName]; !ok {
				tenantsToUsers[tName] = []UserConfig{{
					User:     user.Username,
					Password: decryptedUserPassword,
					Permissions: NatsPermissions{
						Publish:   NatsAuthorization{Allow: allowWriteSubjects},
						Subscribe: NatsAuthorization{Allow: allowReadSubjects},
					},
				}}
			} else {
				tenantsToUsers[tName] = append(usrMap, UserConfig{
					User:     user.Username,
					Password: decryptedUserPassword,
					Permissions: NatsPermissions{
						Publish:   NatsAuthorization{Allow: allowWriteSubjects},
						Subscribe: NatsAuthorization{Allow: allowReadSubjects},
					},
				})
			}
		} else {
			if usrMap, ok := tenantsToUsers[tName]; !ok {
				tenantsToUsers[tName] = []UserConfig{{User: user.Username, Password: decryptedUserPassword}}
			} else {
				tenantsToUsers[tName] = append(usrMap, UserConfig{User: user.Username, Password: decryptedUserPassword})
			}
		}
	}
	for _, t := range tenants {
		decryptedUserPassword, err := DecryptAES(decriptionKey, t.InternalWSPass)
		if err != nil {
			return _EMPTY_, err
		}
		internalAppUser := fmt.Sprintf("$%s$%v", t.Name, t.ID) // for internal use
		usrsList := []UserConfig{{User: "$" + t.Name, Password: getInternalUserPassword()}, {User: internalAppUser, Password: configuration.CONNECTION_TOKEN + "_" + configuration.ROOT_PASSWORD}, {User: MEMPHIS_USERNAME + "$" + strconv.Itoa(t.ID), Password: decryptedUserPassword}}
		if usrMap, ok := tenantsToUsers[t.Name]; ok {
			var usrChangeName UserConfig
			for _, usr := range usrMap {
				usrChangeName = UserConfig{User: usr.User + "$" + strconv.Itoa(t.ID), Password: usr.Password, Permissions: usr.Permissions}
				usrsList = append(usrsList, usrChangeName)
			}
		}
		maxConnAllowed := noLimit
		if IsStorageLimitExceeded(t.Name) {
			maxConnAllowed = 0
		}
		accounts[t.Name] = AccountConfig{Jetstream: &enableJetStream, Users: usrsList, Limits: map[string]*int{"max_connections": &maxConnAllowed}, Imports: memphisReplaceImportString}
	}
	accounts[MEMPHIS_GLOBAL_ACCOUNT] = AccountConfig{Jetstream: &enableJetStream, Users: globalUsers, Limits: map[string]*int{"max_connections": &noLimit}, Exports: memphisReplaceExportString}
	jsonString, err := generateJSONString(accounts)
	if err != nil {
		return _EMPTY_, err
	}
	jsonString = strings.ReplaceAll(jsonString, `"replaceImports"`, memphisImportString)
	jsonString = strings.ReplaceAll(jsonString, `"replaceExports"`, memphisExportString)
	jsonString = strings.ReplaceAll(jsonString, "\\u003e", ">")
	return jsonString, nil
}

func upsertAccountsAndUsers(Accounts []*Account, Users []*User) error {
	if len(Accounts) > 0 {
		tenantsToUpsert := []models.TenantForUpsert{}
		for _, account := range Accounts {
			name := account.GetName()
			if account.GetName() != DEFAULT_SYSTEM_ACCOUNT {
				name = strings.ToLower(name)
				encryptedPass, err := EncryptAES([]byte(generateRandomPassword(12)))
				if err != nil {
					return err
				}
				tenantsToUpsert = append(tenantsToUpsert, models.TenantForUpsert{Name: name, InternalWSPass: encryptedPass})
			}
		}
		err := db.UpsertBatchOfTenants(tenantsToUpsert)
		if err != nil {
			return err
		}
	}
	if len(Users) > 0 {
		usersToUpsert := []models.User{}
		for _, user := range Users {
			if user.Account.GetName() != DEFAULT_SYSTEM_ACCOUNT {
				username := strings.ToLower(user.Username)
				tenantName := strings.ToLower(user.Account.GetName())
				newUser := models.User{
					Username:   username,
					Password:   user.Password,
					UserType:   "application",
					CreatedAt:  time.Now(),
					AvatarId:   1,
					FullName:   _EMPTY_,
					TenantName: tenantName,
				}
				usersToUpsert = append(usersToUpsert, newUser)
			}
		}
		if len(usersToUpsert) > 0 {
			err := db.UpsertBatchOfUsers(usersToUpsert)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) MoveResourcesFromOldToNewDefaultAcc() error {
	stations, err := db.GetAllStations()
	if err != nil {
		return err
	}

	stationsMap := map[int]models.Station{}
	for _, station := range stations {
		stationName, err := StationNameFromStr(station.Name)
		if err != nil {
			return err
		}
		stationsMap[station.ID] = station
		err = s.CreateStream(MEMPHIS_GLOBAL_ACCOUNT, stationName, station.RetentionType, station.RetentionValue, station.StorageType, station.IdempotencyWindow, station.Replicas, station.TieredStorageEnabled, 0, false)
		if err != nil {
			return err
		}
		err = s.RemoveStream(DEFAULT_GLOBAL_ACCOUNT, stationName.Intern())
		if err != nil {
			return err
		}
	}
	consumers, err := db.GetConsumers()
	if err != nil {
		return err
	}
	for _, consumer := range consumers {
		station := stationsMap[consumer.StationId]
		err = s.CreateConsumer(consumer.TenantName, consumer, station, station.PartitionsList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) getIp() string {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		serv.Warnf("getIp: error get ip: %s", err.Error())
		return _EMPTY_
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		serv.Warnf("getIp: error reading response get ip body: %s", err.Error())
		return _EMPTY_
	}
	return string(ip)
}

func (s *Server) SendReloadSignal() error {
	err := s.sendInternalAccountMsgWithReply(serv.MemphisGlobalAccount(), CONFIGURATIONS_RELOAD_SIGNAL_SUBJ, _EMPTY_, nil, _EMPTY_, true)
	return err
}
