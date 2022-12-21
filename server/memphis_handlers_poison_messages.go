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
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"memphis-broker/models"
	"memphis-broker/notifications"
	"sort"
	"strconv"
	"strings"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const PoisonMessageTitle = "Poison message"

type PoisonMessagesHandler struct{ S *Server }

func (s *Server) ListenForPoisonMessages() {
	s.queueSubscribe("$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>",
		"$memphis_poison_messages_listeners_group",
		createPoisonMessageHandler(s))
}

func createPoisonMessageHandler(s *Server) simplifiedMsgHandler {
	return func(_ *client, _, _ string, msg []byte) {
		go s.handleNewPoisonMessage(copyBytes(msg))
	}
}

func (s *Server) handleNewPoisonMessage(msg []byte) {
	var message map[string]interface{}
	err := json.Unmarshal(msg, &message)
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	streamName := message["stream"].(string)
	stationName := StationNameFromStreamName(streamName)
	cgName := message["consumer"].(string)
	cgName = revertDelimiters(cgName)
	messageSeq := message["stream_seq"].(float64)
	deliveriesCount := message["deliveries"].(float64)

	poisonMessageContent, err := s.memphisGetMessage(stationName.Intern(), uint64(messageSeq))
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	_, station, err := IsStationExist(stationName)
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	producerDetails := models.ProducerDetails{}
	producedByHeader := ""
	poisonedCg := models.PoisonedCg{}
	messagePayload := models.MessagePayloadDlq{
		TimeSent: poisonMessageContent.Time,
		Size:     len(poisonMessageContent.Subject) + len(poisonMessageContent.Data) + len(poisonMessageContent.Header),
		Data:     hex.EncodeToString(poisonMessageContent.Data),
	}

	if station.IsNative {
		headersJson, err := DecodeHeader(poisonMessageContent.Header)

		if err != nil {
			serv.Errorf("handleNewPoisonMessage: " + err.Error())
			return
		}
		connectionIdHeader := headersJson["$memphis_connectionId"]
		producedByHeader = headersJson["$memphis_producedBy"]

		//This check for backward compatability
		if connectionIdHeader == "" || producedByHeader == "" {
			connectionIdHeader = headersJson["connectionId"]
			producedByHeader = headersJson["producedBy"]
			if connectionIdHeader == "" || producedByHeader == "" {
				serv.Warnf("handleNewPoisonMessage: Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
				return
			}
		}

		if producedByHeader == "$memphis_dlq" { // skip poison messages which have been resent
			return
		}

		connId, _ := primitive.ObjectIDFromHex(connectionIdHeader)
		_, conn, err := IsConnectionExist(connId)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		filter := bson.M{"name": producedByHeader, "connection_id": connId}
		var producer models.Producer
		err = producersCollection.FindOne(context.TODO(), filter).Decode(&producer)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		producerDetails = models.ProducerDetails{
			Name:          producedByHeader,
			ClientAddress: conn.ClientAddress,
			ConnectionId:  connId,
			CreatedByUser: producer.CreatedByUser,
			IsActive:      producer.IsActive,
			IsDeleted:     producer.IsDeleted,
		}

		poisonedCg = models.PoisonedCg{
			CgName:          cgName,
			PoisoningTime:   time.Now(),
			DeliveriesCount: int(deliveriesCount),
		}

		messagePayload.Headers = headersJson
	}

	id := GetDlqMsgId(stationName.Intern(), int(messageSeq), producedByHeader, poisonMessageContent.Time)
	pmMessage := models.DlqMessage{
		ID:           id,
		StationName:  stationName.Ext(),
		MessageSeq:   int(messageSeq),
		Producer:     producerDetails,
		PoisonedCg:   poisonedCg,
		Message:      messagePayload,
		CreationDate: time.Now(),
	}
	poisonSubjectName := GetDlqSubject("poison", stationName.Intern(), id)
	msgToSend, err := json.Marshal(pmMessage)
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}
	s.sendInternalAccountMsg(s.GlobalAccount(), poisonSubjectName, msgToSend)

	idForUrl := pmMessage.ID
	var msgUrl = idForUrl + "/stations/" + stationName.Ext() + "/" + idForUrl
	err = notifications.SendNotification(PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, notifications.PoisonMAlert)
	if err != nil {
		serv.Warnf("handleNewPoisonMessage: Error while sending a poison message notification: " + err.Error())
		return
	}
}

func (pmh PoisonMessagesHandler) GetDlqMsgsByStationLight(station models.Station) ([]models.LightDlqMessageResponse, []models.LightDlqMessageResponse, error) {
	poisonMessages := make([]models.LightDlqMessageResponse, 0)
	schemaMessages := make([]models.LightDlqMessageResponse, 0)

	timeout := 1 * time.Second

	sn, err := StationNameFromStr(station.Name)
	if err != nil {
		return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
	}

	amount := min(streamInfo.State.Msgs, 1000)
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
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
				serv.Errorf("GetDlqMsgsByStationLight: " + err.Error())
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
		return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
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
	idCheck := make(map[string]bool)
	if err != nil {
		return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]

		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return []models.LightDlqMessageResponse{}, []models.LightDlqMessageResponse{}, err
		}
		msgId := dlqMsg.ID
		if msgType == "poison" {
			if _, value := idCheck[msgId]; !value {
				idCheck[msgId] = true
				poisonMessages = append(poisonMessages, models.LightDlqMessageResponse{MessageSeq: int(msg.Sequence), ID: msgId, Message: dlqMsg.Message})
			}
		} else if msgType == "schema" {
			if _, value := idCheck[msgId]; !value {
				idCheck[msgId] = true
				schemaMessages = append(schemaMessages, models.LightDlqMessageResponse{MessageSeq: int(msg.Sequence), ID: msgId, Message: dlqMsg.Message})
			}
		}
	}

	sort.Slice(poisonMessages, func(i, j int) bool {
		return poisonMessages[i].Message.TimeSent.After(poisonMessages[j].Message.TimeSent)
	})

	sort.Slice(schemaMessages, func(i, j int) bool {
		return schemaMessages[i].Message.TimeSent.After(schemaMessages[j].Message.TimeSent)
	})

	return poisonMessages, schemaMessages, nil
}

func (pmh PoisonMessagesHandler) GetDlqMsgsByStationFull(station models.Station) ([]models.DlqMessageResponse, []models.DlqMessageResponse, error) {
	poisonMessages := make([]models.DlqMessageResponse, 0)
	schemaMessages := make([]models.DlqMessageResponse, 0)
	idCheck := make(map[string]bool)
	idToMsgListP := make(map[string]models.DlqMessageResponse)
	idToMsgListS := make(map[string]models.DlqMessageResponse)

	timeout := 1 * time.Second

	sn, err := StationNameFromStr(station.Name)
	if err != nil {
		return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
	}

	amount := min(streamInfo.State.Msgs, 1000)
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
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
				serv.Errorf("GetDlqMsgsByStationFull: " + err.Error())
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
		return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
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
		return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
	}

	cgToMsgListP := make(map[string]bool)
	cgToMsgListS := make(map[string]bool)
	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage

		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
		}

		msgId := dlqMsg.ID
		if msgType == "poison" {
			if _, value := idCheck[msgId]; !value {
				cgInfo, err := pmh.S.GetCgInfo(sn, dlqMsg.PoisonedCg.CgName)
				if err != nil {
					return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
				}
				pCg := dlqMsg.PoisonedCg
				pCg.UnprocessedMessages = int(cgInfo.NumPending)
				pCg.InProcessMessages = cgInfo.NumAckPending
				pCg.TotalPoisonMessages, err = GetTotalPoisonMsgsByCg(dlqMsg.StationName, dlqMsg.PoisonedCg.CgName)
				if err != nil {
					return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
				}
				idCheck[msgId] = true
				idToMsgListP[msgId] = models.DlqMessageResponse{
					ID:           msgId,
					StationName:  dlqMsg.StationName,
					MessageSeq:   dlqMsg.MessageSeq,
					Producer:     dlqMsg.Producer,
					Message:      dlqMsg.Message,
					CreationDate: dlqMsg.CreationDate,
					PoisonedCgs:  []models.PoisonedCg{pCg},
				}
			} else {
				if _, value := cgToMsgListP[dlqMsg.PoisonedCg.CgName]; !value {
					cgInfo, err := pmh.S.GetCgInfo(sn, dlqMsg.PoisonedCg.CgName)
					if err != nil {
						return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
					}
					pCg := dlqMsg.PoisonedCg
					pCg.UnprocessedMessages = int(cgInfo.NumPending)
					pCg.InProcessMessages = cgInfo.NumAckPending
					pCg.TotalPoisonMessages, err = GetTotalPoisonMsgsByCg(dlqMsg.StationName, dlqMsg.PoisonedCg.CgName)
					if err != nil {
						return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
					}
					cgToMsgListP[dlqMsg.PoisonedCg.CgName] = true
					dlm := idToMsgListP[msgId]
					pcgs := append(dlm.PoisonedCgs, pCg)
					dlm.PoisonedCgs = pcgs
					idToMsgListP[msgId] = dlm
				}
			}
		} else if msgType == "schema" {
			if _, value := idCheck[msgId]; !value {
				cgInfo, err := pmh.S.GetCgInfo(sn, dlqMsg.PoisonedCg.CgName)
				if err != nil {
					return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
				}
				pCg := dlqMsg.PoisonedCg
				pCg.UnprocessedMessages = int(cgInfo.NumPending)
				pCg.InProcessMessages = cgInfo.NumAckPending
				pCg.TotalPoisonMessages, err = GetTotalPoisonMsgsByCg(dlqMsg.StationName, dlqMsg.PoisonedCg.CgName)
				if err != nil {
					return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
				}
				idCheck[msgId] = true
				idToMsgListS[msgId] = models.DlqMessageResponse{
					ID:           msgId,
					StationName:  dlqMsg.StationName,
					MessageSeq:   dlqMsg.MessageSeq,
					Producer:     dlqMsg.Producer,
					Message:      dlqMsg.Message,
					CreationDate: dlqMsg.CreationDate,
					PoisonedCgs:  []models.PoisonedCg{pCg},
				}
			} else {
				if _, value := cgToMsgListS[dlqMsg.PoisonedCg.CgName]; !value {
					cgInfo, err := pmh.S.GetCgInfo(sn, dlqMsg.PoisonedCg.CgName)
					if err != nil {
						return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
					}
					pCg := dlqMsg.PoisonedCg
					pCg.UnprocessedMessages = int(cgInfo.NumPending)
					pCg.InProcessMessages = cgInfo.NumAckPending
					pCg.TotalPoisonMessages, err = GetTotalPoisonMsgsByCg(dlqMsg.StationName, dlqMsg.PoisonedCg.CgName)
					if err != nil {
						return []models.DlqMessageResponse{}, []models.DlqMessageResponse{}, err
					}
					cgToMsgListS[dlqMsg.PoisonedCg.CgName] = true
					dlm := idToMsgListS[msgId]
					dlm.PoisonedCgs = append(dlm.PoisonedCgs, pCg)
					idToMsgListS[msgId] = dlm
				}
			}
		}
	}

	for _, dlmRes := range idToMsgListP {
		poisonMessages = append(poisonMessages, dlmRes)
	}

	for _, dlmRes := range idToMsgListS {
		schemaMessages = append(schemaMessages, dlmRes)
	}

	sort.Slice(poisonMessages, func(i, j int) bool {
		return poisonMessages[i].Message.TimeSent.After(poisonMessages[j].Message.TimeSent)
	})

	sort.Slice(schemaMessages, func(i, j int) bool {
		return schemaMessages[i].Message.TimeSent.After(schemaMessages[j].Message.TimeSent)
	})

	return poisonMessages, schemaMessages, nil
}

func (pmh PoisonMessagesHandler) GetTotalPoisonMsgsByStation(stationName string) (int, error) {
	count := 0
	timeout := 1 * time.Second
	idCheck := make(map[string]bool)

	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return 0, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return 0, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return 0, err
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
				serv.Errorf("GetTotalPoisonMsgsByCg: " + err.Error())
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
		return 0, err
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
		return 0, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return 0, err
		}
		msgId := dlqMsg.ID
		if msgType == "poison" {
			if _, value := idCheck[msgId]; !value {
				count++
			}
		}
	}

	return count, nil
}

func RemovePoisonedCg(stationName StationName, cgName string) error {
	timeout := 1 * time.Second

	streamName := fmt.Sprintf(dlsStreamName, stationName.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return err
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
				serv.Errorf("GetTotalPoisonMsgsByCg: " + err.Error())
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
		return err
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
		return err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return err
		}
		if msgType == "poison" {
			if dlqMsg.PoisonedCg.CgName == cgName {
				_, err = serv.memphisDeleteMsgFromStream(streamName, msg.Sequence)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func GetTotalPoisonMsgsByCg(stationName, cgName string) (int, error) {
	count := 0
	timeout := 1 * time.Second

	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return 0, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return 0, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return 0, err
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
				serv.Errorf("GetTotalPoisonMsgsByCg: " + err.Error())
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
		return 0, err
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
		return 0, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return 0, err
		}
		if msgType == "poison" {
			if dlqMsg.PoisonedCg.CgName == cgName {
				count++
			}
		}
	}

	return count, nil
}

func GetTotalSchemaFailMsgsByCg(stationName, cgName string) (int, error) {
	count := 0
	timeout := 1 * time.Second
	idCheck := make(map[string]bool)

	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return 0, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dlqp_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return 0, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return 0, err
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
				serv.Errorf("GetTotalPoisonMsgsByCg: " + err.Error())
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
		return 0, err
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
		return 0, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return 0, err
		}
		msgId := dlqMsg.ID
		if msgType == "schema" {
			if _, value := idCheck[msgId]; !value {
				idCheck[msgId] = true
				count++
			}
		}
	}

	return count, nil
}

func GetPoisonedCgsByMessage(stationNameInter string, message models.MessageDetails) ([]models.PoisonedCg, error) {
	timeout := 1 * time.Second
	poisonedCgs := []models.PoisonedCg{}
	streamName := fmt.Sprintf(dlsStreamName, stationNameInter)
	cgCheck := make(map[string]bool)
	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_pcg_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}
	msgId := GetDlqMsgId(stationNameInter, message.MessageSeq, message.ProducedBy, message.TimeSent)
	filter := GetDlqSubject("poison", stationNameInter, msgId)
	cc := ConsumerConfig{
		OptStartSeq:   startSeq,
		DeliverPolicy: DeliverByStartSequence,
		AckPolicy:     AckExplicit,
		Durable:       durableName,
		FilterSubject: filter,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return []models.PoisonedCg{}, err
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
				serv.Errorf("GetTotalPoisonMsgsByCg: " + err.Error())
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
		return []models.PoisonedCg{}, err
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
		return []models.PoisonedCg{}, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlqMsg models.DlqMessage
		err = json.Unmarshal(msg.Data, &dlqMsg)
		if err != nil {
			return []models.PoisonedCg{}, err
		}
		if msgType == "poison" {
			if _, value := cgCheck[dlqMsg.PoisonedCg.CgName]; !value {
				cgCheck[dlqMsg.PoisonedCg.CgName] = true
				poisonedCgs = append(poisonedCgs, dlqMsg.PoisonedCg)
			}
		}
	}

	return poisonedCgs, nil
}

func GetDlqSubject(subjType string, stationName string, id string) string {
	return fmt.Sprintf(dlsStreamName, stationName) + "." + subjType + "." + id
}

func GetDlqMsgId(stationName string, messageSeq int, producerName string, timeSent time.Time) string {
	return strings.ReplaceAll(stationName+"-"+producerName+"-"+strconv.Itoa(messageSeq)+"-"+timeSent.String(), " ", "")
}

func GetDlqSubject(subjType string, stationName string, id string) string {
	return fmt.Sprintf(dlsStreamName, stationName) + "." + subjType + "." + id
}

func GetPoisonMsgId(messageSeq int, producerName string, timeSent time.Time) string {
	return producerName + "-" + timeSent.String() + "-" + strconv.Itoa(messageSeq)
}
