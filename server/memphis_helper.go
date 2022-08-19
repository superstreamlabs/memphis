package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"memphis-broker/models"
	"net/textproto"
	"strings"
	"time"
)

const (
	crlf      = "\r\n"
	hdrPreEnd = len(hdrLine) - len(crlf)
	statusLen = 3 // e.g. 20x, 40x, 50x
	statusHdr = "Status"
	descrHdr  = "Description"
)

// internal reply subjects
const (
	replySubjectStreamInfo     = "$memphis_stream_info_reply"
	replySubjectCreateConsumer = "$memphis_create_consumer_reply"
	replySubjectDeleteConsumer = "$memphis_delete_consumer_reply"
	replySubjectConsumerInfo   = "$memphis_consumer_info_reply"
	replySubjectDeleteStream   = "$memphis_delete_stream_reply"
	replySubjectStreamList     = "$memphis_stream_list_reply"
	replySubjectGetMsg         = "$memphis_get_msg_reply"
)

// errors
var (
	ErrBadHeader = errors.New("could not decode header")
)

func (s *Server) MemphisInitialized() bool {
	return s.GlobalAccount().JetStreamEnabled()
}

func (s *Server) jsApiRequest(subject, reply string, msg []byte) []byte {
	// signal the handler that we will be waiting for a reply
	s.memphis.replySubjectActive[reply] = true

	// send on golbal account
	s.sendInternalAccountMsgWithReply(s.GlobalAccount(), subject, reply, nil, msg, true)

	// wait for response to arrive
	rawResp := <-s.memphis.replySubjectRespCh[reply]

	s.memphis.replySubjectActive[reply] = false

	return rawResp
}

func AddUser(username string) (string, error) {
	return configuration.CONNECTION_TOKEN, nil
}

func RemoveUser(username string) error {
	return nil
}

func (s *Server) CreateStream(station models.Station) error {
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

	return s.memphisAddStream(&StreamConfig{
		Name:         station.Name,
		Subjects:     []string{station.Name + ".>"},
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

func (s *Server) memphisAddStream(sc *StreamConfig) error {
	acc := s.GlobalAccount()
	_, err := acc.addStream(sc)

	return err
}

func (s *Server) CreateConsumer(consumer models.Consumer, station models.Station) error {
	var consumerName string
	if consumer.ConsumersGroup != "" {
		consumerName = consumer.ConsumersGroup
	} else {
		consumerName = consumer.Name
	}

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

	err := s.memphisAddConsumer(station.Name, &ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: DeliverAll,
		AckPolicy:     AckExplicit,
		AckWait:       time.Duration(maxAckTimeMs) * time.Millisecond,
		MaxDeliver:    MaxMsgDeliveries,
		FilterSubject: station.Name + ".final",
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

	rawResp := s.jsApiRequest(requestSubject, replySubjectCreateConsumer, []byte(rawRequest))

	var resp JSApiConsumerCreateResponse
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("ConsumerCreate json response unmarshal error")
		return err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) RemoveConsumer(streamName string, cn string) error {
	requestSubject := fmt.Sprintf(JSApiConsumerDeleteT, streamName, cn)

	rawResp := s.jsApiRequest(requestSubject, replySubjectDeleteConsumer, []byte(_EMPTY_))

	var resp JSApiConsumerDeleteResponse
	err := json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("ConsumerDelete json response unmarshal error")
		return err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GetCgInfo(stationName, cgName string) (*ConsumerInfo, error) {
	requestSubject := fmt.Sprintf(JSApiConsumerInfoT, stationName, cgName)

	rawResp := s.jsApiRequest(requestSubject, replySubjectConsumerInfo, []byte(_EMPTY_))

	var resp JSApiConsumerInfoResponse
	err := json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("ConsumerInfo json response unmarshal error")
		return nil, err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return nil, err
	}

	return resp.ConsumerInfo, nil
}

func (s *Server) RemoveStream(streamName string) error {
	requestSubject := fmt.Sprintf(JSApiStreamDeleteT, streamName)

	rawResp := s.jsApiRequest(requestSubject, replySubjectDeleteStream, []byte(_EMPTY_))

	var resp JSApiStreamDeleteResponse
	err := json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("StreamDelete json response unmarshal error")
		return err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GetTotalMessagesInStation(station models.Station) (int, error) {
	streamInfo, err := s.memphisStreamInfo(station.Name)
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

func (s *Server) memphisStreamInfo(streamName string) (*StreamInfo, error) {

	requestSubject := fmt.Sprintf(JSApiStreamInfoT, streamName)

	rawResp := s.jsApiRequest(requestSubject, replySubjectStreamInfo, []byte(_EMPTY_))

	var resp JSApiStreamInfoResponse
	err := json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("StreamInfo json response unmarshal error")
		return nil, err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return nil, err
	}

	return resp.StreamInfo, nil
}

func (s *Server) GetAvgMsgSizeInStation(station models.Station) (int64, error) {
	streamInfo, err := s.memphisStreamInfo(station.Name)
	if err != nil || streamInfo.State.Bytes == 0 {
		return 0, err
	}

	return int64(streamInfo.State.Bytes / streamInfo.State.Msgs), nil
}

func (s *Server) memphisAllStreamsInfo() ([]*StreamInfo, error) {
	requestSubject := fmt.Sprintf(JSApiStreamList)

	request := JSApiStreamListRequest{}
	rawRequest, err := json.Marshal(request)
	rawResp := s.jsApiRequest(requestSubject, replySubjectStreamList, []byte(rawRequest))

	var resp JSApiStreamListResponse
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("StreamList json response unmarshal error")
		return nil, err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return nil, err
	}

	return resp.Streams, nil
}

func (s *Server) GetMessages(station models.Station, messagesToFetch int) ([]models.MessageDetails, error) {
	streamInfo, err := s.memphisStreamInfo(station.Name)
	if err != nil {
		return []models.MessageDetails{}, err
	}
	totalMessages := streamInfo.State.Msgs

	var startSequence uint64 = 1
	if totalMessages > uint64(messagesToFetch) {
		startSequence = totalMessages - uint64(messagesToFetch) + 1
	} else {
		messagesToFetch = int(totalMessages)
	}

	msgs, err := s.memphisGetMsgs(station.Name+".final",
		station.Name,
		startSequence,
		messagesToFetch,
		3*time.Second)
	var messages []models.MessageDetails
	if err != nil {
		return []models.MessageDetails{}, err
	}

	for _, msg := range msgs {
		hdr, err := DecodeHeader(msg.Header)
		if err != nil {
			return nil, err
		}
		if hdr["producedBy"] == "$memphis_dlq" { // skip poison messages which have been resent
			continue
		}

		data := (string(msg.Data))
		if len(data) > 100 { // get the first chars for preview needs
			data = data[0:100]
		}
		messages = append(messages, models.MessageDetails{
			MessageSeq:   int(msg.Sequence),
			Data:         data,
			ProducedBy:   hdr["producedBy"],
			ConnectionId: hdr["connectionId"],
			TimeSent:     msg.Time,
			Size:         len(msg.Subject) + len(msg.Data) + len(msg.Header),
		})
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 { // sort from new to old
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (s *Server) memphisGetMsgs(subjectName,
	streamName string,
	startSeq uint64,
	amount int,
	timeout time.Duration) ([]StoredMsg, error) {
	var msgs []StoredMsg
	seq := startSeq

	timeoutCh := time.After(timeout)
	for i := 0; i < amount; i++ {
		select {
		case <-timeoutCh:
			return msgs, errors.New("MemphisGetMsgs timeout")
		default:
			sm, err := s.GetMessage(streamName, seq)
			if sm == nil || err != nil {
				return msgs, err
			}
			seq = sm.Sequence + 1
			msgs = append(msgs, *sm)
		}
	}

	return msgs, nil
}

func (s *Server) GetMessage(streamName string, msgSeq uint64) (*StoredMsg, error) {
	requestSubject := fmt.Sprintf(JSApiMsgGetT, streamName)

	request := JSApiMsgGetRequest{Seq: msgSeq}

	rawRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	rawResp := s.jsApiRequest(requestSubject, replySubjectGetMsg, rawRequest)

	var resp JSApiMsgGetResponse
	err = json.Unmarshal(rawResp, &resp)
	if err != nil {
		s.Errorf("MsgGet json response unmarshal error")
		return nil, err
	}

	err = resp.ApiResponse.ToError()
	if err != nil {
		return nil, err
	}

	return resp.Message, nil
}

func (s *Server) queueSubscribe(subj, queueGroupName string, cb func(string, []byte)) error {
	acc := s.GlobalAccount()
	c := acc.ic
	wcb := func(_ *subscription, _ *client, _ *Account, subject, _ string, rmsg []byte) {
		cb(subject, rmsg)
	}

	_, err := c.processSub([]byte(subj), []byte(queueGroupName), []byte("memphis_internal"), wcb, false)

	return err
}

func (s *Server) subscribeOnGlobalAcc(subj, sid string, cb func(string, string, []byte)) (*subscription, error) {
	acc := s.GlobalAccount()
	c := acc.ic
	wcb := func(_ *subscription, _ *client, _ *Account, subject, reply string, rmsg []byte) {
		cb(subject, reply, rmsg)
	}

	return c.processSub([]byte(subj), nil, []byte(sid), wcb, false)
}

func (s *Server) Respond(reply string, msg []byte) {
	acc := s.GlobalAccount()
	s.sendInternalAccountMsg(acc, reply, msg)
}

func (s *Server) ResendPoisonMessage(subject string, data []byte) error {
	hdr := map[string]string{"producedBy": "$memphis_dlq"}
	s.sendInternalMsgWithHeaderLocked(subject, hdr, data)
	return nil
}

func (s *Server) sendInternalMsgWithHeaderLocked(subj string, hdr map[string]string, msg interface{}) {
	s.mu.Lock()
	if s.sys == nil || s.sys.sendq == nil {
		return
	}
	s.sys.sendq.push(newPubMsg(nil, subj, _EMPTY_, nil, hdr, msg, noCompression, false, false))
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
