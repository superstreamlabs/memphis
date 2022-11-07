// Credit for The NATS.IO Authors
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
// limitations under the License.package server
package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"memphis-broker/models"
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
	syslogsStreamName  = "$memphis_syslogs"
	syslogsInfoSubject = "info"
	syslogsWarnSubject = "warn"
	syslogsErrSubject  = "err"
)

// JetStream API request kinds
const (
	kindStreamInfo     = "$memphis_stream_info"
	kindCreateConsumer = "$memphis_create_consumer"
	kindDeleteConsumer = "$memphis_delete_consumer"
	kindConsumerInfo   = "$memphis_consumer_info"
	kindCreateStream   = "$memphis_create_stream"
	kindDeleteStream   = "$memphis_delete_stream"
	kindStreamList     = "$memphis_stream_list"
	kindGetMsg         = "$memphis_get_msg"
)

// errors
var (
	ErrBadHeader = errors.New("could not decode header")
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
		sub.close()
	case <-timeout:
		sub.close()
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
	return configuration.CONNECTION_TOKEN, nil
}

func RemoveUser(username string) error {
	return nil
}

func (s *Server) CreateStream(sn StationName, station models.Station) error {
	var maxMsgs int
	if station.RetentionType == "messages" && station.RetentionValue > 0 {
		maxMsgs = station.RetentionValue
	} else {
		maxMsgs = -1
	}

	var maxBytes int
	if station.RetentionType == "bytes" && station.RetentionValue > 0 {
		maxBytes = station.RetentionValue
	} else {
		maxBytes = -1
	}

	var maxAge time.Duration
	if station.RetentionType == "message_age_sec" && station.RetentionValue > 0 {
		maxAge = time.Duration(station.RetentionValue) * time.Second
	} else {
		maxAge = time.Duration(0)
	}

	var storage StorageType
	if station.StorageType == "memory" {
		storage = MemoryStorage
	} else {
		storage = FileStorage
	}

	var dedupWindow time.Duration
	if station.DedupEnabled && station.DedupWindowInMs >= 100 {
		dedupWindow = time.Duration(station.DedupWindowInMs) * time.Millisecond
	} else {
		dedupWindow = time.Duration(100) * time.Millisecond // can not be 0
	}

	return s.
		memphisAddStream(&StreamConfig{
			Name:         sn.Intern(),
			Subjects:     []string{sn.Intern() + ".>"},
			Retention:    LimitsPolicy,
			MaxConsumers: -1,
			MaxMsgs:      int64(maxMsgs),
			MaxBytes:     int64(maxBytes),
			Discard:      DiscardOld,
			MaxAge:       maxAge,
			MaxMsgsPer:   -1,
			MaxMsgSize:   int32(configuration.MAX_MESSAGE_SIZE_MB) * 1024 * 1024,
			Storage:      storage,
			Replicas:     station.Replicas,
			NoAck:        false,
			Duplicates:   dedupWindow,
		})
}

func (s *Server) memphisClusterReady() {
	if !s.memphis.mcrReported {
		s.memphis.mcrReported = true
		close(s.memphis.mcr)
	}
}

func (s *Server) CreateSystemLogsStream() {
	if s.JetStreamIsClustered() {
		if !s.memphis.logStreamCreated {
			timeout := time.NewTimer(2 * time.Minute)
			select {
			case <-timeout.C:
				s.Fatalf("Failed to create syslogs stream: cluster readiness timeout")
			case <-s.memphis.mcr:
				timeout.Stop()
				s.memphis.logStreamCreated = true
			}

			if !s.JetStreamIsLeader() {
				return
			}
		}

	}

	retentionDays, err := strconv.Atoi(configuration.LOGS_RETENTION_IN_DAYS)
	if err != nil {
		s.Fatalf("Failed to create syslogs stream: " + " " + err.Error())

	}
	retentionDur := time.Duration(retentionDays) * time.Hour * 24

	err = s.memphisAddStream(&StreamConfig{
		Name:         syslogsStreamName,
		Subjects:     []string{syslogsStreamName + ".>"},
		Retention:    LimitsPolicy,
		MaxAge:       retentionDur,
		MaxConsumers: -1,
		Discard:      DiscardOld,
		Storage:      FileStorage,
	})
	if err != nil {
		s.Fatalf("Failed to create syslogs stream: " + " " + err.Error())
	}

	if s.memphis.activateSysLogsPubFunc == nil {
		s.Fatalf("publish activation func is not initialised")
	}
	s.memphis.activateSysLogsPubFunc()
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

	err = s.memphisAddConsumer(stationName.Intern(), &ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckExplicit,
		AckWait:       time.Duration(maxAckTimeMs) * time.Millisecond,
		MaxDeliver:    MaxMsgDeliveries,
		FilterSubject: stationName.Intern() + ".final",
		ReplayPolicy:  ReplayInstant,
		MaxAckPending: -1,
		HeadersOnly:   false,
		// RateLimit: ,// Bits per sec
		// Heartbeat: // time.Duration,
	})
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

func (s *Server) GetTotalMessagesInStation(stationName StationName) (int, error) {
	streamInfo, err := s.memphisStreamInfo(stationName.Intern())
	if err != nil {
		return 0, err
	}

	return int(streamInfo.State.Msgs), nil
}

func (s *Server) GetTotalMessagesAcrossAllStations() (int, error) {
	messagesCounter := 0

	streams, err := s.memphisAllStreamsInfo()
	if err != nil {
		return messagesCounter, err
	}

	for _, streamInfo := range streams {
		if !strings.HasPrefix(streamInfo.Config.Name, "$memphis") { // skip internal streams
			messagesCounter = messagesCounter + int(streamInfo.State.Msgs)
		}
	}

	return messagesCounter, nil
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

	request := JSApiStreamListRequest{}
	rawRequest, err := json.Marshal(request)
	var resp JSApiStreamListResponse
	err = jsApiRequest(s, requestSubject, kindStreamList, []byte(rawRequest), &resp)
	if err != nil {
		return nil, err
	}

	err = resp.ToError()
	if err != nil {
		return nil, err
	}

	return resp.Streams, nil
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

	msgs, err := s.memphisGetMsgs(stationName.Intern()+".final",
		stationName.Intern(),
		startSequence,
		messagesToFetch,
		5*time.Second)
	var messages []models.MessageDetails
	if err != nil {
		return []models.MessageDetails{}, err
	}

	for _, msg := range msgs {
		headersJson, err := DecodeHeader(msg.Header)
		if err != nil {
			return nil, err
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

		for key, _ := range headersJson {
			if strings.HasPrefix(key, "$memphis") {
				delete(headersJson, key)
			}
		}

		// delete(headersJson, "$memphis_connectionId")
		// delete(headersJson, "$memphis_producedBy")

		if producedByHeader == "$memphis_dlq" { // skip poison messages which have been resent
			continue
		}

		data := (string(msg.Data))
		if len(data) > 100 { // get the first chars for preview needs
			data = data[0:100]
		}

		// headersJson, err := getMessageHeaders(hdr)
		// if err != nil {
		// 	return []models.MessageDetails{}, err
		// }
		messages = append(messages, models.MessageDetails{
			MessageSeq:   int(msg.Sequence),
			Data:         data,
			ProducedBy:   producedByHeader,
			ConnectionId: connectionIdHeader,
			TimeSent:     msg.Time,
			Size:         len(msg.Subject) + len(msg.Data) + len(msg.Header),
			Headers:      headersJson,
		})
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

func (s *Server) memphisGetMsgs(subjectName, streamName string, startSeq uint64, amount int, timeout time.Duration) ([]StoredMsg, error) {
	uid, _ := uuid.NewV4()
	durableName := "$memphis_fetch_messages_consumer_" + uid.String()

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		Durable:       durableName,
		AckPolicy:     AckExplicit,
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
		go func(respCh chan StoredMsg, reply string, msg []byte) {
			// ack
			s.sendInternalAccountMsg(s.GlobalAccount(), reply, []byte(_EMPTY_))

			rawTs := tokenAt(reply, 8)
			seq, _, _ := ackReplyInfo(reply)

			intTs, err := strconv.Atoi(rawTs)
			if err != nil {
				s.Errorf(err.Error())
			}

			dataFirstIdx := getHdrLastIdxFromRaw(msg) + 1
			if dataFirstIdx == 0 || dataFirstIdx > len(msg)-len(CR_LF) {
				s.Errorf("memphis error parsing in station get messages")
			}

			dataLen := len(msg) - dataFirstIdx

			respCh <- StoredMsg{
				Sequence: uint64(seq),
				Header:   msg[:dataFirstIdx],
				Data:     msg[dataFirstIdx : dataFirstIdx+dataLen],
				Time:     time.Unix(0, int64(intTs)),
			}
		}(responseChan, reply, copyBytes(msg))
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
	sub.close()
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

	hdrs["$memphis_producedBy"] = "$memphis_dlq"

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
	// mh, err := readMIMEHeader(tp)
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
