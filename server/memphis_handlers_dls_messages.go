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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"memphis/db"
	"memphis/models"
	"sort"
	"strconv"
	"strings"

	"time"
)

const (
	PoisonMessageTitle = "Poison message"
	dlsMsgSep          = "~"
)

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
	_, station, err := db.GetStationByName(stationName.Ext())
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}
	if !station.DlsConfigurationPoison {
		return
	}

	cgName := message["consumer"].(string)
	cgName = revertDelimiters(cgName)
	messageSeq := message["stream_seq"].(float64)
	deliveriesCount := message["deliveries"].(float64)

	poisonMessageContent, err := s.memphisGetMessage(stationName.Intern(), uint64(messageSeq))
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	producerDetails := models.ProducerDetails{}
	producedByHeader := ""
	poisonedCg := models.PoisonedCg{}

	var headersJson map[string]string
	if poisonMessageContent.Header != nil {
		headersJson, err = DecodeHeader(poisonMessageContent.Header)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: " + err.Error())
			return
		}
	}
	messagePayload := models.MessagePayloadDls{
		TimeSent: poisonMessageContent.Time,
		Size:     len(poisonMessageContent.Subject) + len(poisonMessageContent.Data) + len(poisonMessageContent.Header),
		Data:     hex.EncodeToString(poisonMessageContent.Data),
		Headers:  headersJson,
	}

	if station.IsNative {
		connectionIdHeader := headersJson["$memphis_connectionId"]
		producedByHeader = headersJson["$memphis_producedBy"]

		// This check for backward compatability
		if connectionIdHeader == "" || producedByHeader == "" {
			connectionIdHeader = headersJson["connectionId"]
			producedByHeader = headersJson["producedBy"]
			if connectionIdHeader == "" || producedByHeader == "" {
				serv.Warnf("handleNewPoisonMessage: Error while getting notified about a poison message: Missing mandatory message headers, please upgrade the SDK version you are using")
				return
			}
		}

		if producedByHeader == "$memphis_dls" { // skip poison messages which have been resent
			return
		}

		connId := connectionIdHeader
		_, conn, err := db.GetConnectionByID(connId)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}
		exist, producer, err := db.GetProducerByNameAndConnectionID(producedByHeader, connId)
		if !exist {
			serv.Warnf("handleNewPoisonMessage: producer " + producedByHeader + " couldn't been found")
			return
		}
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		producerDetails = models.ProducerDetails{
			Name:          producedByHeader,
			ClientAddress: conn.ClientAddress,
			ConnectionId:  connId,
			CreatedBy:     producer.CreatedBy,
			IsActive:      producer.IsActive,
			IsDeleted:     producer.IsDeleted,
		}

		poisonedCg = models.PoisonedCg{
			CgName:          cgName,
			PoisoningTime:   time.Now(),
			DeliveriesCount: int(deliveriesCount),
		}
	}

	id := GetDlsMsgId(stationName.Intern(), int(messageSeq), producedByHeader, poisonMessageContent.Time.String())
	pmMessage := models.DlsMessage{
		ID:          id,
		StationName: stationName.Ext(),
		MessageSeq:  int(messageSeq),
		Producer:    producerDetails,
		PoisonedCg:  poisonedCg,
		Message:     messagePayload,
		CreatedAt:   time.Now(),
	}
	internalCgName := replaceDelimiters(cgName)
	poisonSubjectName := GetDlsSubject("poison", stationName.Intern(), id, internalCgName)
	msgToSend, err := json.Marshal(pmMessage)
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}
	s.sendInternalAccountMsg(s.GlobalAccount(), poisonSubjectName, msgToSend)

	idForUrl := pmMessage.ID
	var msgUrl = UI_HOST + "/stations/" + stationName.Ext() + "/" + idForUrl
	err = SendNotification(PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, PoisonMAlert)
	if err != nil {
		serv.Warnf("handleNewPoisonMessage: Error while sending a poison message notification: " + err.Error())
		return
	}
}

func (pmh PoisonMessagesHandler) GetDlsMsgsByStationLight(station models.Station) ([]models.LightDlsMessageResponse, []models.LightDlsMessageResponse, int, map[string]int, error) {
	poisonMessages := make([]models.LightDlsMessageResponse, 0)
	schemaMessages := make([]models.LightDlsMessageResponse, 0)
	poisonedCgMap := make(map[string]int)
	timeout := 1 * time.Second

	sn, err := StationNameFromStr(station.Name)
	if err != nil {
		return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dls_consumer_" + uid
	var msgs []StoredMsg

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
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
		Replicas:      1,
	}

	err = serv.memphisAddConsumer(streamName, &cc)
	if err != nil {
		return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
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
				serv.Errorf("GetDlsMsgsByStationLight: " + err.Error())
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
		return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
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
		return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
	}

	for _, msg := range msgs {
		splittedSubj := strings.Split(msg.Subject, tsep)
		msgType := splittedSubj[1]
		var dlsMsg models.DlsMessage
		err = json.Unmarshal(msg.Data, &dlsMsg)
		if err != nil {
			return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, poisonedCgMap, err
		}
		msgId := dlsMsg.ID
		if msgType == "poison" {
			if _, ok := idCheck[msgId]; !ok {
				idCheck[msgId] = true
				poisonMessages = append(poisonMessages, models.LightDlsMessageResponse{MessageSeq: int(msg.Sequence), ID: msgId, Message: dlsMsg.Message})
			}
			if _, ok := poisonedCgMap[dlsMsg.PoisonedCg.CgName]; !ok {
				poisonedCgMap[dlsMsg.PoisonedCg.CgName] = 1
			} else {
				poisonedCgMap[dlsMsg.PoisonedCg.CgName] += 1
			}
		} else {
			if _, value := idCheck[msgId]; !value {
				idCheck[msgId] = true
				message := dlsMsg.Message
				if dlsMsg.CreatedAt.IsZero() {
					message.TimeSent = time.Unix(0, dlsMsg.CreationUnix*1000000)
				} else {
					message.TimeSent = dlsMsg.CreatedAt
				}
				message.Size = len(msg.Subject) + len(message.Data) + len(message.Headers)
				schemaMessages = append(schemaMessages, models.LightDlsMessageResponse{MessageSeq: int(msg.Sequence), ID: msgId, Message: message})
			}
		}
	}

	lenPoison, lenSchema := len(poisonMessages), len(schemaMessages)
	totalDlsAmount := lenPoison + lenSchema

	sort.Slice(poisonMessages, func(i, j int) bool {
		return poisonMessages[i].Message.TimeSent.After(poisonMessages[j].Message.TimeSent)
	})

	sort.Slice(schemaMessages, func(i, j int) bool {
		return schemaMessages[i].Message.TimeSent.After(schemaMessages[j].Message.TimeSent)
	})

	if lenPoison > 1000 {
		poisonMessages = poisonMessages[:1000]
	}

	if lenSchema > 1000 {
		schemaMessages = schemaMessages[:1000]
	}
	return poisonMessages, schemaMessages, totalDlsAmount, poisonedCgMap, nil
}

func getDlsMessageById(station models.Station, sn StationName, dlsMsgId, dlsType string) (models.DlsMessageResponse, error) {
	timeout := 500 * time.Millisecond
	dlsStreamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	streamInfo, err := serv.memphisStreamInfo(dlsStreamName)
	if err != nil {
		return models.DlsMessageResponse{}, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}

	dlsType = strings.ToLower(dlsType)
	var filterSubj string
	switch dlsType {
	case "poison":
		filterSubj = GetDlsSubject(dlsType, sn.Intern(), dlsMsgId, ">")
	case "schema":
		filterSubj = GetDlsSubject(dlsType, sn.Intern(), dlsMsgId, "")
	default:
		filterSubj = GetDlsSubject("poison", sn.Intern(), dlsMsgId, ">")
	}

	msgs, err := serv.memphisGetMessagesByFilter(dlsStreamName, filterSubj, startSeq, amount, timeout)
	if err != nil {
		return models.DlsMessageResponse{}, err
	}

	if len(msgs) < 1 {
		serv.Warnf("getDlsMessageById: no dls message with id: %s", dlsMsgId)
		return models.DlsMessageResponse{}, err
	}

	msgType := tokenAt(msgs[0].Subject, 2)

	// if msgType == "poison"
	poisonedCgs := []models.PoisonedCg{}
	var producer models.Producer
	var dlsMsg models.DlsMessage
	var clientAddress string
	var connectionId string

	for i, msg := range msgs {
		err = json.Unmarshal(msg.Data, &dlsMsg)
		if err != nil {
			return models.DlsMessageResponse{}, err
		}

		if station.IsNative {
			if i == 0 {
				connectionIdHeader := dlsMsg.Message.Headers["$memphis_connectionId"]
				//This check for backward compatability
				if connectionIdHeader == "" {
					connectionIdHeader = dlsMsg.Message.Headers["connectionId"]
					if connectionIdHeader == "" {
						return models.DlsMessageResponse{}, err
					}
				}
				connectionId = connectionIdHeader
				_, conn, err := db.GetConnectionByID(connectionId)
				if err != nil {
					return models.DlsMessageResponse{}, err
				}
				clientAddress = conn.ClientAddress
				exist, prod, err := db.GetProducerByNameAndConnectionID(dlsMsg.Producer.Name, connectionId)
				if err != nil {
					return models.DlsMessageResponse{}, err
				}
				if !exist {
					return models.DlsMessageResponse{}, errors.New("Producer " + dlsMsg.Producer.Name + " does not exist")
				}
				producer = prod
			}

			if msgType == "poison" {
				cgInfo, err := serv.GetCgInfo(sn, dlsMsg.PoisonedCg.CgName)
				if err != nil {
					return models.DlsMessageResponse{}, err
				}

				pCg := dlsMsg.PoisonedCg
				pCg.UnprocessedMessages = int(cgInfo.NumPending)
				pCg.InProcessMessages = cgInfo.NumAckPending
				cgMembers, err := GetConsumerGroupMembers(pCg.CgName, station)
				if err != nil {
					return models.DlsMessageResponse{}, err
				}
				pCg.IsActive, pCg.IsDeleted = getCgStatus(cgMembers)

				pCg.TotalPoisonMessages = -1
				pCg.MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
				pCg.MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
				pCg.CgMembers = cgMembers
				poisonedCgs = append(poisonedCgs, pCg)
			}

			if msgType == "schema" {
				size := len(msg.Subject) + len(dlsMsg.Message.Data) + len(dlsMsg.Message.Headers)
				dlsMsg.Message.Size = size
				if dlsMsg.CreatedAt.IsZero() {
					dlsMsg.Message.TimeSent = time.Unix(0, dlsMsg.CreationUnix*1000000)
				} else {
					dlsMsg.Message.TimeSent = dlsMsg.CreatedAt
				}
			}

			for header := range dlsMsg.Message.Headers {
				if strings.HasPrefix(header, "$memphis") {
					delete(dlsMsg.Message.Headers, header)
				}
			}
		}
	}

	sort.Slice(poisonedCgs, func(i, j int) bool {
		return poisonedCgs[i].PoisoningTime.After(poisonedCgs[j].PoisoningTime)
	})

	schemaType := ""
	if station.SchemaName != "" {
		exist, schema, err := db.GetSchemaByName(station.SchemaName)
		if err != nil {
			return models.DlsMessageResponse{}, err
		}
		if exist {
			schemaType = schema.Type
		}
	}

	result := models.DlsMessageResponse{
		ID:          dlsMsgId,
		StationName: dlsMsg.StationName,
		SchemaType:  schemaType,
		MessageSeq:  dlsMsg.MessageSeq,
		Producer: models.ProducerDetails{
			Name:              producer.Name,
			ConnectionId:      producer.ConnectionId,
			ClientAddress:     clientAddress,
			CreatedBy:         producer.CreatedBy,
			CreatedByUsername: producer.CreatedByUsername,
			IsActive:          producer.IsActive,
			IsDeleted:         producer.IsDeleted,
		},
		Message:         dlsMsg.Message,
		CreatedAt:       dlsMsg.CreatedAt,
		PoisonedCgs:     poisonedCgs,
		ValidationError: dlsMsg.ValidationError,
	}

	return result, nil
}

func (pmh PoisonMessagesHandler) GetTotalDlsMsgsByStation(stationName string) (int, error) {
	count := 0
	timeout := 1 * time.Second
	idCheck := make(map[string]bool)

	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return 0, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dls_consumer_" + uid
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
		Replicas:      1,
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
		var dlsMsg models.DlsMessage
		err = json.Unmarshal(msg.Data, &dlsMsg)
		if err != nil {
			return 0, err
		}
		msgId := dlsMsg.ID
		if msgType == "poison" {
			if _, ok := idCheck[msgId]; !ok {
				idCheck[msgId] = true
				count++
			}
		} else if msgType == "schema" {
			count++
		}
	}

	return count, nil
}

func RemovePoisonedCg(stationName StationName, cgName string) error {
	timeout := 500 * time.Millisecond

	streamName := fmt.Sprintf(dlsStreamName, stationName.Intern())

	uid := serv.memphis.nuid.Next()
	durableName := "$memphis_fetch_dls_consumer_" + uid
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
		Replicas:      1,
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
		var dlsMsg models.DlsMessage
		err = json.Unmarshal(msg.Data, &dlsMsg)
		if err != nil {
			return err
		}
		if msgType == "poison" {
			if dlsMsg.PoisonedCg.CgName == cgName {
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
	timeout := 500 * time.Millisecond

	sn, err := StationNameFromStr(stationName)
	if err != nil {
		return 0, err
	}
	streamName := fmt.Sprintf(dlsStreamName, sn.Intern())

	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return 0, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}
	internalCgName := replaceDelimiters(cgName)
	filter := GetDlsSubject("poison", sn.Intern(), "*", internalCgName)
	msgs, err := serv.memphisGetMessagesByFilter(streamName, filter, startSeq, amount, timeout)
	if err != nil {
		return 0, err
	}
	return len(msgs), nil
}

func GetPoisonedCgsByMessage(stationNameInter string, message models.MessageDetails) ([]models.PoisonedCg, error) {
	timeout := 500 * time.Millisecond
	poisonedCgs := []models.PoisonedCg{}
	streamName := fmt.Sprintf(dlsStreamName, stationNameInter)
	streamInfo, err := serv.memphisStreamInfo(streamName)
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	amount := streamInfo.State.Msgs
	startSeq := uint64(1)
	if streamInfo.State.FirstSeq > 0 {
		startSeq = streamInfo.State.FirstSeq
	}
	msgId := GetDlsMsgId(stationNameInter, message.MessageSeq, message.ProducedBy, message.TimeSent.String())
	filter := GetDlsSubject("poison", stationNameInter, msgId, "*")
	msgs, err := serv.memphisGetMessagesByFilter(streamName, filter, 0, amount, timeout)
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	if uint64(len(msgs)) < amount && streamInfo.State.Msgs > amount && streamInfo.State.FirstSeq < startSeq {
		return GetPoisonedCgsByMessage(stationNameInter, message)
	}

	for _, msg := range msgs {
		var dlsMsg models.DlsMessage
		err = json.Unmarshal(msg.Data, &dlsMsg)
		if err != nil {
			return []models.PoisonedCg{}, err
		}

		poisonedCgs = append(poisonedCgs, dlsMsg.PoisonedCg)
	}

	sort.Slice(poisonedCgs, func(i, j int) bool {
		return poisonedCgs[i].PoisoningTime.After(poisonedCgs[j].PoisoningTime)
	})

	return poisonedCgs, nil
}

func GetDlsSubject(subjType, stationName, id, cgName string) string {
	suffix := _EMPTY_
	if cgName != _EMPTY_ {
		suffix = tsep + cgName
	}
	return fmt.Sprintf(dlsStreamName, stationName) + tsep + subjType + tsep + id + suffix
}

func GetDlsMsgId(stationName string, messageSeq int, producerName string, timeSent string) string {
	producer := producerName
	// Support for dls messages from nonNative Memphis SDKs
	if producer == "" {
		producer = "nonNative"
	}
	// Remove any spaces might be in ID
	msgId := strings.ReplaceAll(stationName+dlsMsgSep+producer+dlsMsgSep+strconv.Itoa(messageSeq)+dlsMsgSep+timeSent, " ", "")
	msgId = strings.ReplaceAll(msgId, tsep, "+")
	return msgId
}
