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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"memphis/models"
	"net/textproto"
	"sort"
	"strconv"
	"strings"
	"time"

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
	syslogsStreamName      = "$memphis_syslogs"
	syslogsExternalSubject = "extern.*"
	syslogsInfoSubject     = "extern.info"
	syslogsWarnSubject     = "extern.warn"
	syslogsErrSubject      = "extern.err"
	syslogsSysSubject      = "intern.sys"
	dlsStreamName          = "$memphis-%s-dls"
	tieredStorageStream    = "$memphis_tiered_storage"
	throughputStreamName   = "$memphis-throughput"
	throughputStreamNameV1 = "$memphis-throughput-v1"
)

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
)

// errors
var (
	ErrBadHeader                    = errors.New("could not decode header")
	LOGS_RETENTION_IN_DAYS          int
	DLS_RETENTION_HOURS             int
	TIERED_STORAGE_CONSUMER_CREATED bool
	TIERED_STORAGE_STREAM_CREATED   bool
	BROKER_HOST                     string
	UI_HOST                         string
	REST_GW_HOST                    string
	TIERED_STORAGE_TIME_FRAME_SEC   int
)

func (s *Server) MemphisInitialized() bool {
	return s.GlobalAccount().JetStreamEnabled()
}

func createReplyHandler(s *Server, respCh chan []byte) simplifiedMsgHandler {
	return func(_ *client, subject, _ string, msg []byte) {
		go func(msg []byte) {
			respCh <- msg
		}(copyBytes(msg))
	}
}

func jsApiRequest[R any](s *Server, subject, kind string, msg []byte, resp *R) error {
	reply := s.getJsApiReplySubject()

	s.memphis.jsApiMu.Lock()
	defer s.memphis.jsApiMu.Unlock()

	timeout := time.After(30 * time.Second)
	respCh := make(chan []byte)
	sub, err := s.subscribeOnGlobalAcc(reply, reply+"_sid", createReplyHandler(s, respCh))
	if err != nil {
		return err
	}
	// send on global account
	s.sendInternalAccountMsgWithReply(s.GlobalAccount(), subject, reply, nil, msg, true)

	// wait for response to arrive
	var rawResp []byte
	select {
	case rawResp = <-respCh:
		s.unsubscribeOnGlobalAcc(sub)
		break
	case <-timeout:
		s.unsubscribeOnGlobalAcc(sub)
		return fmt.Errorf("jsapi request timeout for request type %q on %q", kind, subject)
	}

	return json.Unmarshal(rawResp, resp)
}

func (s *Server) getJsApiReplySubject() string {
	var sb strings.Builder
	sb.WriteString("$memphis_jsapi_reply_")
	sb.WriteString(nuid.Next())
	return sb.String()
}

func AddUser(username string) (string, error) {
	return serv.opts.Authorization, nil
}

func RemoveUser(username string) error {
	return nil
}

func (s *Server) CreateStream(sn StationName, retentionType string, retentionValue int, storageType string, idempotencyW int64, replicas int, tieredStorageEnabled bool) error {
	var maxMsgs int
	if retentionType == "messages" && retentionValue > 0 {
		maxMsgs = retentionValue
	} else {
		maxMsgs = -1
	}

	var maxBytes int
	if retentionType == "bytes" && retentionValue > 0 {
		maxBytes = retentionValue
	} else {
		maxBytes = -1
	}

	var maxAge time.Duration
	if retentionType == "message_age_sec" && retentionValue > 0 {
		maxAge = time.Duration(retentionValue) * time.Second
	} else {
		maxAge = time.Duration(0)
	}

	var storage StorageType
	if storageType == "memory" {
		storage = MemoryStorage
	} else {
		storage = FileStorage
	}

	var idempotencyWindow time.Duration
	if idempotencyW <= 0 {
		idempotencyWindow = 2 * time.Minute // default
	} else if idempotencyW < 100 {
		idempotencyWindow = time.Duration(100) * time.Millisecond // minimum is 100 millis
	} else {
		idempotencyWindow = time.Duration(idempotencyW) * time.Millisecond
	}

	return s.
		memphisAddStream(&StreamConfig{
			Name:                 sn.Intern(),
			Subjects:             []string{sn.Intern() + ".>"},
			Retention:            LimitsPolicy,
			MaxConsumers:         -1,
			MaxMsgs:              int64(maxMsgs),
			MaxBytes:             int64(maxBytes),
			Discard:              DiscardOld,
			MaxAge:               maxAge,
			MaxMsgsPer:           -1,
			Storage:              storage,
			Replicas:             replicas,
			NoAck:                false,
			Duplicates:           idempotencyWindow,
			TieredStorageEnabled: tieredStorageEnabled,
		})
}

func (s *Server) CreateDlsStream(sn StationName, storageType string, replicas int) error {
	maxAge := time.Duration(DLS_RETENTION_HOURS) * time.Hour

	var storage StorageType
	if storageType == "memory" {
		storage = MemoryStorage
	} else {
		storage = FileStorage
	}

	idempotencyWindow := time.Duration(100) * time.Millisecond // minimum is 100 millis

	name := fmt.Sprintf(dlsStreamName, sn.Intern())

	return s.
		memphisAddStream(&StreamConfig{
			Name:         (name),
			Subjects:     []string{name + ".>"},
			Retention:    LimitsPolicy,
			MaxConsumers: -1,
			MaxMsgs:      int64(-1),
			MaxBytes:     int64(-1),
			Discard:      DiscardOld,
			MaxAge:       maxAge,
			MaxMsgsPer:   -1,
			Storage:      storage,
			Replicas:     replicas,
			NoAck:        false,
			Duplicates:   idempotencyWindow,
		})
}

func (s *Server) CreateInternalJetStreamResources() {
	ready := !s.JetStreamIsClustered()
	retentionDur := time.Duration(LOGS_RETENTION_IN_DAYS) * time.Hour * 24

	successCh := make(chan error)

	if ready { // stand alone
		go tryCreateInternalJetStreamResources(s, retentionDur, successCh, false)
		err := <-successCh
		if err != nil {
			s.Errorf("CreateInternalJetStreamResources: system streams creation failed: " + err.Error())
		}
	} else {
		for !ready { // wait for cluster to be ready if we are in cluster mode
			timeout := time.NewTimer(1 * time.Minute)
			go tryCreateInternalJetStreamResources(s, retentionDur, successCh, true)
			select {
			case <-timeout.C:
				s.Warnf("CreateInternalJetStreamResources: system streams creation takes more than a minute")
				err := <-successCh
				if err != nil {
					s.Warnf("CreateInternalJetStreamResources: " + err.Error())
					continue
				}
				ready = true
			case err := <-successCh:
				if err != nil {
					s.Warnf("CreateInternalJetStreamResources: " + err.Error())
					<-timeout.C
					continue
				}
				timeout.Stop()
				ready = true
			}
		}
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
	err = s.memphisAddStream(&StreamConfig{
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
	if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
		successCh <- err
		return
	}

	if s.memphis.activateSysLogsPubFunc == nil {
		s.Fatalf("internal error: sys logs publish activation func is not initialized")
	}
	s.memphis.activateSysLogsPubFunc()
	s.popFallbackLogs()

	idempotencyWindow := time.Duration(1 * time.Minute)
	// tiered storage stream
	err = s.memphisAddStream(&StreamConfig{
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
	if err != nil && !IsNatsErr(err, JSStreamNameExistErr) {
		successCh <- err
		return
	}
	TIERED_STORAGE_STREAM_CREATED = true

	// create tiered storage consumer
	durableName := TIERED_STORAGE_CONSUMER
	tieredStorageTimeFrame := time.Duration(TIERED_STORAGE_TIME_FRAME_SEC) * time.Second
	filterSubject := tieredStorageStream + ".>"
	cc := ConsumerConfig{
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
		FilterSubject: filterSubject,
		AckWait:       time.Duration(2) * tieredStorageTimeFrame,
		MaxAckPending: -1,
		MaxDeliver:    1,
	}
	err = serv.memphisAddConsumer(tieredStorageStream, &cc)
	if err != nil {
		successCh <- err
		return
	}
	TIERED_STORAGE_CONSUMER_CREATED = true

	// delete the old version throughput stream
	err = s.memphisDeleteStream(throughputStreamName)
	if err != nil && !IsNatsErr(err, JSStreamNotFoundErr) {
		s.Errorf("Failed deleting old internal throughput stream - %s", err.Error())

	}

	// throughput kv
	err = s.memphisAddStream(&StreamConfig{
		Name:         (throughputStreamNameV1),
		Subjects:     []string{throughputStreamNameV1 + ".>"},
		Retention:    LimitsPolicy,
		MaxConsumers: -1,
		MaxMsgs:      int64(-1),
		MaxBytes:     int64(-1),
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

func (s *Server) memphisAddStream(sc *StreamConfig) error {
	requestSubject := fmt.Sprintf(JSApiStreamCreateT, sc.Name)

	request, err := json.Marshal(sc)
	if err != nil {
		return err
	}

	var resp JSApiStreamCreateResponse
	err = jsApiRequest(s, requestSubject, kindCreateStream, request, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) memphisDeleteStream(streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamDeleteT, streamName)

	var resp JSApiStreamCreateResponse
	err := jsApiRequest(s, requestSubject, kindCreateStream, nil, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) memphisUpdateStream(sc *StreamConfig) error {
	requestSubject := fmt.Sprintf(JSApiStreamUpdateT, sc.Name)

	request, err := json.Marshal(sc)
	if err != nil {
		return err
	}

	var resp JSApiStreamUpdateResponse
	err = jsApiRequest(s, requestSubject, kindUpdateStream, request, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func getInternalConsumerName(cn string) string {
	return replaceDelimiters(cn)
}

func (s *Server) CreateConsumer(consumer models.Consumer, station models.Station) error {
	var consumerName string
	if consumer.ConsumersGroup != "" {
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

	var deliveryPolicy DeliverPolicy
	streamInfo, err := serv.memphisStreamInfo(stationName.Intern())
	if err != nil {
		return errors.New("Streaminfo: " + err.Error())
	}
	lastSeq := streamInfo.State.LastSeq

	var optStartSeq uint64
	// This check for case when the last message is 0 (in case StartConsumeFromSequence > 1 the LastMessages is 0 )
	if consumer.LastMessages == 0 && consumer.StartConsumeFromSeq == 0 {
		deliveryPolicy = DeliverNew
	} else if consumer.LastMessages > 0 {
		lastMessages := (lastSeq - uint64(consumer.LastMessages)) + 1
		if int(lastMessages) < 1 {
			lastMessages = uint64(1)
		}
		deliveryPolicy = DeliverByStartSequence
		optStartSeq = lastMessages
	} else if consumer.StartConsumeFromSeq == 1 || consumer.LastMessages == -1 {
		deliveryPolicy = DeliverAll
	} else if consumer.StartConsumeFromSeq > 1 {
		deliveryPolicy = DeliverByStartSequence
		optStartSeq = consumer.StartConsumeFromSeq
	}

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
	err = s.memphisAddConsumer(stationName.Intern(), consumerConfig)
	return err
}

func (s *Server) memphisAddConsumer(streamName string, cc *ConsumerConfig) error {
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
	err = jsApiRequest(s, requestSubject, kindCreateConsumer, []byte(rawRequest), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) RemoveConsumer(stationName StationName, cn string) error {
	cn = getInternalConsumerName(cn)
	return s.memphisRemoveConsumer(stationName.Intern(), cn)
}

func (s *Server) memphisRemoveConsumer(streamName, cn string) error {
	requestSubject := fmt.Sprintf(JSApiConsumerDeleteT, streamName, cn)
	var resp JSApiConsumerDeleteResponse
	err := jsApiRequest(s, requestSubject, kindDeleteConsumer, []byte(_EMPTY_), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) GetCgInfo(stationName StationName, cgName string) (*ConsumerInfo, error) {
	cgName = replaceDelimiters(cgName)
	requestSubject := fmt.Sprintf(JSApiConsumerInfoT, stationName.Intern(), cgName)

	var resp JSApiConsumerInfoResponse
	err := jsApiRequest(s, requestSubject, kindConsumerInfo, []byte(_EMPTY_), &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.ConsumerInfo, nil
}

func (s *Server) RemoveStream(streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamDeleteT, streamName)

	var resp JSApiStreamDeleteResponse
	err := jsApiRequest(s, requestSubject, kindDeleteStream, []byte(_EMPTY_), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) PurgeStream(streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamPurgeT, streamName)

	var resp JSApiStreamPurgeResponse
	err := jsApiRequest(s, requestSubject, kindPurgeStream, []byte(_EMPTY_), &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) Opts() *Options {
	return s.opts
}

func (s *Server) AnalyticsToken() string {
	return ANALYTICS_TOKEN
}

func (s *Server) MemphisVersion() string {
	return VERSION
}

func (s *Server) RemoveMsg(stationName StationName, msgSeq uint64) error {
	requestSubject := fmt.Sprintf(JSApiMsgDeleteT, stationName.Intern())

	var resp JSApiMsgDeleteResponse
	req := JSApiMsgDeleteRequest{Seq: msgSeq}
	reqj, _ := json.Marshal(req)
	err := jsApiRequest(s, requestSubject, kindDeleteMessage, reqj, &resp)
	if err != nil {
		return err
	}

	return resp.ToError()
}

func (s *Server) GetTotalMessagesInStation(stationName StationName) (int, error) {
	streamInfo, err := s.memphisStreamInfo(stationName.Intern())
	if err != nil {
		return 0, err
	}

	return int(streamInfo.State.Msgs), nil
}

// low level call, call only with internal station name (i.e stream name)!
func (s *Server) memphisStreamInfo(streamName string) (*StreamInfo, error) {
	requestSubject := fmt.Sprintf(JSApiStreamInfoT, streamName)

	var resp JSApiStreamInfoResponse
	err := jsApiRequest(s, requestSubject, kindStreamInfo, []byte(_EMPTY_), &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.StreamInfo, nil
}

func (s *Server) memphisDeleteMsgFromStream(streamName string, seq uint64) (ApiResponse, error) {
	requestSubject := fmt.Sprintf(JSApiMsgDeleteT, streamName)

	msg := JSApiMsgDeleteRequest{
		Seq: seq,
	}

	req, err := json.Marshal(msg)
	if err != nil {
		return ApiResponse{}, err
	}

	var resp JSApiMsgDeleteResponse
	err = jsApiRequest(s, requestSubject, kindDeleteMsg, req, &resp)
	if err != nil {
		return ApiResponse{}, err
	}

	err = resp.ToError()
	if err != nil {
		return ApiResponse{}, err
	}

	return resp.ApiResponse, nil
}

func (s *Server) GetAvgMsgSizeInStation(station models.Station) (int64, error) {
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return 0, err
	}

	streamInfo, err := s.memphisStreamInfo(stationName.Intern())
	if err != nil || streamInfo.State.Bytes == 0 {
		return 0, err
	}

	return int64(streamInfo.State.Bytes / streamInfo.State.Msgs), nil
}

func (s *Server) memphisAllStreamsInfo() ([]*StreamInfo, error) {
	requestSubject := fmt.Sprintf(JSApiStreamList)
	streams := make([]*StreamInfo, 0)

	offset := 0
	offsetReq := ApiPagedRequest{Offset: offset}
	request := JSApiStreamListRequest{ApiPagedRequest: offsetReq}
	rawRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	var resp JSApiStreamListResponse
	err = jsApiRequest(s, requestSubject, kindStreamList, []byte(rawRequest), &resp)
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

		err = jsApiRequest(s, requestSubject, kindStreamList, []byte(rawRequest), &resp)
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
	streamInfo, err := s.memphisStreamInfo(stationName.Intern())
	if err != nil {
		return []models.MessageDetails{}, err
	}
	totalMessages := streamInfo.State.Msgs
	lastStreamSeq := streamInfo.State.LastSeq

	var startSequence uint64 = 1
	if totalMessages > uint64(messagesToFetch) {
		startSequence = lastStreamSeq - uint64(messagesToFetch) + 1
	} else {
		messagesToFetch = int(totalMessages)
	}

	filterSubj := stationName.Intern() + ".final"
	if !station.IsNative {
		filterSubj = ""
	}

	msgs, err := s.memphisGetMsgs(filterSubj,
		stationName.Intern(),
		startSequence,
		messagesToFetch,
		5*time.Second,
		true,
	)
	var messages []models.MessageDetails
	if err != nil {
		return []models.MessageDetails{}, err
	}

	stationIsNative := station.IsNative

	for _, msg := range msgs {
		messageDetails := models.MessageDetails{
			MessageSeq: int(msg.Sequence),
			TimeSent:   msg.Time,
			Size:       len(msg.Subject) + len(msg.Data) + len(msg.Header),
		}

		data := hex.EncodeToString(msg.Data)
		if len(data) > 40 { // get the first chars for preview needs
			data = data[0:40]
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

			//This check for backward compatability
			if connectionIdHeader == "" || producedByHeader == "" {
				connectionIdHeader = headersJson["connectionId"]
				producedByHeader = strings.ToLower(headersJson["producedBy"])
				if connectionIdHeader == "" || producedByHeader == "" {
					return []models.MessageDetails{}, errors.New("Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
				}
			}

			for header := range headersJson {
				if strings.HasPrefix(header, "$memphis") {
					delete(headersJson, header)
				}
			}

			if producedByHeader == "$memphis_dls" { // skip poison messages which have been resent
				continue
			}
			messageDetails.ProducedBy = producedByHeader
			messageDetails.ConnectionId = connectionIdHeader
			messageDetails.Headers = headersJson
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

func (s *Server) memphisGetMsgs(filterSubj, streamName string, startSeq uint64, amount int, timeout time.Duration, findHeader bool) ([]StoredMsg, error) {
	uid, _ := uuid.NewV4()
	durableName := "$memphis_fetch_messages_consumer_" + uid.String()

	cc := ConsumerConfig{
		FilterSubject: filterSubj,
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		Durable:       durableName,
		AckPolicy:     AckExplicit,
		Replicas:      1,
	}

	err := s.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return nil, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, streamName, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.Itoa(amount))

	sub, err := s.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, reply string, msg []byte, findHeader bool) {
			// ack
			s.sendInternalAccountMsg(s.GlobalAccount(), reply, []byte(_EMPTY_))

			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf("memphisGetMsgs: " + err.Error())
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

	s.sendInternalAccountMsgWithReply(s.GlobalAccount(), subject, reply, nil, req, true)

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
	s.unsubscribeOnGlobalAcc(sub)
	err = s.memphisRemoveConsumer(streamName, durableName)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (s *Server) GetMessage(stationName StationName, msgSeq uint64) (*StoredMsg, error) {
	return s.memphisGetMessage(stationName.Intern(), msgSeq)
}

func (s *Server) GetLeaderAndFollowers(station models.Station) (string, []string, error) {
	var followers []string
	stationName, err := StationNameFromStr(station.Name)
	if err != nil {
		return "", followers, err
	}

	streamInfo, err := s.memphisStreamInfo(stationName.Intern())
	if err != nil {
		return "", followers, err
	}

	for _, replica := range streamInfo.Cluster.Replicas {
		followers = append(followers, replica.Name)
	}

	return streamInfo.Cluster.Leader, followers, nil
}

func (s *Server) memphisGetMessage(streamName string, msgSeq uint64) (*StoredMsg, error) {
	requestSubject := fmt.Sprintf(JSApiMsgGetT, streamName)

	request := JSApiMsgGetRequest{Seq: msgSeq}

	rawRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var resp JSApiMsgGetResponse
	err = jsApiRequest(s, requestSubject, kindGetMsg, rawRequest, &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.Message, nil
}

func (s *Server) memphisGetMessagesByFilter(streamName, filterSubject string, startSeq, amount uint64, timeout time.Duration) ([]StoredMsg, error) {
	uid := serv.memphis.nuid.Next()
	durableName := uid

	deliverPolicy := DeliverAll
	if startSeq != 0 {
		deliverPolicy = DeliverByStartSequence
	}
	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: deliverPolicy,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
		FilterSubject: filterSubject,
		Replicas:      1,
	}
	var msgs []StoredMsg
	err := serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return msgs, err
	}

	responseChan := make(chan StoredMsg)
	subject := fmt.Sprintf(JSApiRequestNextT, streamName, durableName)
	reply := durableName + "_reply"
	req := []byte(strconv.FormatUint(amount, 10))
	sub, err := serv.subscribeOnGlobalAcc(reply, reply+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(respCh chan StoredMsg, subject, reply string, msg []byte) {
			// ack
			serv.sendInternalAccountMsg(serv.GlobalAccount(), reply, []byte(_EMPTY_))
			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				serv.Errorf("dropSchemaDlsMsg: " + err.Error())
			}

			respCh <- StoredMsg{
				Subject:  subject,
				Sequence: uint64(seq),
				Data:     msg,
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, subject, reply, copyBytes(msg))
	})
	if err != nil {
		return msgs, err
	}

	serv.sendInternalAccountMsgWithReply(serv.GlobalAccount(), subject, reply, nil, req, true)

	timer := time.NewTimer(timeout)
	for i := uint64(0); i < amount; i++ {
		select {
		case <-timer.C:
			goto cleanup
		case msg := <-responseChan:
			msgs = append(msgs, msg)
		}
	}

cleanup:
	timer.Stop()
	serv.unsubscribeOnGlobalAcc(sub)
	err = serv.memphisRemoveConsumer(streamName, durableName)
	if err != nil {
		return msgs, err
	}
	return msgs, nil
}

func (s *Server) queueSubscribe(subj, queueGroupName string, cb simplifiedMsgHandler) error {
	acc := s.GlobalAccount()
	c := acc.ic

	acc.mu.Lock()
	acc.isid++
	sid := strconv.FormatUint(acc.isid, 10)
	acc.mu.Unlock()

	wcb := func(_ *subscription, c *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(c, subject, reply, rmsg)
	}

	_, err := c.processSub([]byte(subj), []byte(queueGroupName), []byte(sid), wcb, false)

	return err
}

func (s *Server) subscribeOnGlobalAcc(subj, sid string, cb simplifiedMsgHandler) (*subscription, error) {
	acc := s.GlobalAccount()
	c := acc.ic
	wcb := func(_ *subscription, c *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(c, subject, reply, rmsg)
	}

	return c.processSub([]byte(subj), nil, []byte(sid), wcb, false)
}

func (s *Server) subscribeOnAcc(acc *Account, subj, sid string, cb simplifiedMsgHandler) (*subscription, error) {
	c := acc.ic
	wcb := func(_ *subscription, c *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(c, subject, reply, rmsg)
	}

	return c.processSub([]byte(subj), nil, []byte(sid), wcb, false)
}

func (s *Server) unsubscribeOnGlobalAcc(sub *subscription) error {
	acc := s.GlobalAccount()
	c := acc.ic
	return c.processUnsub(sub.sid)
}

func (s *Server) unsubscribeOnAcc(acc *Account, sub *subscription) error {
	c := acc.ic
	return c.processUnsub(sub.sid)
}

func (s *Server) respondOnGlobalAcc(reply string, msg []byte) {
	acc := s.GlobalAccount()
	s.sendInternalAccountMsg(acc, reply, msg)
}

func (s *Server) ResendPoisonMessage(subject string, data, headers []byte) error {
	hdrs := make(map[string]string)
	err := json.Unmarshal(headers, &hdrs)
	if err != nil {
		return err
	}

	hdrs["$memphis_producedBy"] = "$memphis_dls"

	if hdrs["producedBy"] != "" {
		delete(hdrs, "producedBy")
	}

	s.sendInternalMsgWithHeaderLocked(s.GlobalAccount(), subject, hdrs, data)
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

	hdr := make(map[string]string)
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
		i := bytes.IndexByte([]byte(kv), ':')
		if i < 0 {
			return nil, ErrBadHeader
		}
		key := kv[:i]
		if key == "" {
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
