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
	"context"
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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func UpdatePoisonCgsInDlsMessage(poisonedCgs string, stationId, messageSeq int, updatedAt time.Time) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `UPDATE dls_messages SET poisoned_cgs = ARRAY_APPEND(poisoned_cgs, $1), updated_at = $4 WHERE station_id=$2 AND message_seq=$3`
	stmt, err := conn.Conn().Prepare(ctx, "update_poisoned_cgs", query)
	if err != nil {
		return err
	}
	_, err = conn.Conn().Query(ctx, stmt.Name, poisonedCgs, stationId, messageSeq, updatedAt)
	if err != nil {
		return err
	}
	return nil

}

func GetMsgByStationIdAndMsgSeq(stationId, messageSeq int) (bool, models.DlsMessagePg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	defer connection.Release()

	query := `SELECT * FROM dls_messages WHERE station_id = $1 AND message_seq = $2 LIMIT 1`

	stmt, err := connection.Conn().Prepare(ctx, "get_dls_messages_by_station_id_and_message_seq", query)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq)
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	defer rows.Close()

	message, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessagePg])
	if err != nil {
		return false, models.DlsMessagePg{}, err
	}
	if len(message) == 0 {
		return false, models.DlsMessagePg{}, nil
	}

	return true, message[0], nil

}

func InsertPoisonedCgMessages(stationId int, messageSeq int, producerId int, poisonedCgs []string, messageDetails models.MessagePayloadPg, updatedAt time.Time, messageType string) (models.DlsMessagePg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.DlsMessagePg{}, err
	}
	defer connection.Release()

	query := `INSERT INTO dls_messages( 
			station_id,
			message_seq,
			producer_id,
			poisoned_cgs,
			message_details,
			updated_at,
			message_type
			) 
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	stmt, err := connection.Conn().Prepare(ctx, "insert_dls_messages", query)
	if err != nil {
		return models.DlsMessagePg{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, stationId, messageSeq, producerId, poisonedCgs, messageDetails, updatedAt, messageType)
	if err != nil {
		return models.DlsMessagePg{}, err
	}
	defer rows.Close()
	var messagePaylodId int
	for rows.Next() {
		err := rows.Scan(&messagePaylodId)
		if err != nil {
			return models.DlsMessagePg{}, err
		}
	}

	if err != nil {
		return models.DlsMessagePg{}, err
	}

	msgDetails := models.MessagePayloadDlsPg{
		TimeSent: messageDetails.TimeSent,
		Size:     messageDetails.Size,
		Data:     messageDetails.Data,
		// Headers:  messageDetails.headersJson,
	}

	deadLetterPayload := models.DlsMessagePg{
		ID:             messagePaylodId,
		StationId:      stationId,
		MessageSeq:     messageSeq,
		ProducerId:     producerId,
		PoisonedCgs:    poisonedCgs,
		MessageDetails: msgDetails,
		UpdatedAt:      updatedAt,
	}

	if err := rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Detail != "" {
				if !strings.Contains(pgErr.Detail, "already exists") {
					return models.DlsMessagePg{}, errors.New("messages table already exists")
				} else {
					return models.DlsMessagePg{}, errors.New(pgErr.Detail)
				}
			} else {
				return models.DlsMessagePg{}, errors.New(pgErr.Message)
			}
		} else {
			return models.DlsMessagePg{}, err
		}
	}

	return deadLetterPayload, nil
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

	poisonMessageContent, err := s.memphisGetMessage(stationName.Intern(), uint64(messageSeq))
	if err != nil {
		serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
		return
	}

	producedByHeader := ""

	var headersJson map[string]string
	if poisonMessageContent.Header != nil {
		headersJson, err = DecodeHeader(poisonMessageContent.Header)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: " + err.Error())
			return
		}
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
		_, _, err := db.GetConnectionByID(connId)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		exist, p, err := db.GetProducerByNameAndConnectionID(producedByHeader, connId)
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		if !exist {
			serv.Warnf("handleNewPoisonMessage: producer " + producedByHeader + " couldn't been found")
			return
		}

		updatedAt := time.Now()
		var poisonedCgs []string
		poisonedCgs = append(poisonedCgs, cgName)

		messageDetails := models.MessagePayloadPg{
			TimeSent: poisonMessageContent.Time,
			Size:     len(poisonMessageContent.Data) + len(poisonMessageContent.Header),
			Data:     string(poisonMessageContent.Data),
			Headers:  headersJson,
		}

		exist, deadLetterMsg, err := GetMsgByStationIdAndMsgSeq(station.ID, int(messageSeq))
		if err != nil {
			serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
			return
		}

		if exist {
			err := UpdatePoisonCgsInDlsMessage(cgName, station.ID, int(messageSeq), updatedAt)
			if err != nil {
				serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
				return
			}
		} else {
			deadLetterMsg, err = InsertPoisonedCgMessages(station.ID, int(messageSeq), p.ID, poisonedCgs, messageDetails, updatedAt, "poison")
			if err != nil {
				serv.Errorf("handleNewPoisonMessage: Error while getting notified about a poison message: " + err.Error())
				return
			}
		}

		idForUrl := string(rune(deadLetterMsg.ID))
		var msgUrl = UI_HOST + "/stations/" + stationName.Ext() + "/" + idForUrl
		err = SendNotification(PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, PoisonMAlert)
		if err != nil {
			serv.Warnf("handleNewPoisonMessage: Error while sending a poison message notification: " + err.Error())
			return
		}
	}
}

func (pmh PoisonMessagesHandler) GetDlsMsgsByStationLight(station models.Station) ([]models.LightDlsMessageResponsePg, []models.LightDlsMessageResponsePg, int, error) {
	poisonMessages := make([]models.LightDlsMessageResponsePg, 0)
	schemaMessages := make([]models.LightDlsMessageResponsePg, 0)

	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.LightDlsMessageResponsePg{}, []models.LightDlsMessageResponsePg{}, 0, err
	}
	defer conn.Release()
	query := `SELECT * from dls_messages where station_id=$1 ORDER BY updated_at DESC limit 1000`
	stmt, err := conn.Conn().Prepare(ctx, "get_dls_msg_by_station", query)
	if err != nil {
		return []models.LightDlsMessageResponsePg{}, []models.LightDlsMessageResponsePg{}, 0, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, station.ID)
	if err != nil {
		return []models.LightDlsMessageResponsePg{}, []models.LightDlsMessageResponsePg{}, 0, err
	}
	defer rows.Close()
	dlsMsgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessagePg])
	if err != nil {
		return []models.LightDlsMessageResponsePg{}, []models.LightDlsMessageResponsePg{}, 0, err
	}
	if len(dlsMsgs) == 0 {
		return []models.LightDlsMessageResponsePg{}, []models.LightDlsMessageResponsePg{}, 0, nil
	}

	for _, v := range dlsMsgs {
		switch v.MessageType {
		case "poison":
			messageDetails := models.MessagePayloadDlsPg{
				TimeSent: v.MessageDetails.TimeSent,
				Size:     v.MessageDetails.Size,
				Data:     hex.EncodeToString([]byte(v.MessageDetails.Data)),
				Headers:  v.MessageDetails.Headers,
			}
			poisonMessages = append(poisonMessages, models.LightDlsMessageResponsePg{MessageSeq: v.MessageSeq, ID: v.ID, Message: messageDetails})
		case "schema":
			// message.Size = len(msg.Subject) + len(message.Data) + len(message.Headers)
			schemaMessages = append(schemaMessages, models.LightDlsMessageResponsePg{MessageSeq: v.MessageSeq, ID: v.ID, Message: v.MessageDetails})
		}

	}

	// 	if msgType == "poison" {
	// poisonMessages = append(poisonMessages, models.LightDlsMessageResponse{MessageSeq: , ID: msgId, Message: dlsMsg.Message})
	// 	} else {
	// message.Size = len(msg.Subject) + len(message.Data) + len(message.Headers)
	// schemaMessages = append(schemaMessages, models.LightDlsMessageResponse{MessageSeq: int(msg.Sequence), ID: msgId, Message: message})
	// 		}
	// 	}
	// }

	lenPoison, lenSchema := len(poisonMessages), len(schemaMessages)
	totalDlsAmount := lenPoison + lenSchema

	sort.Slice(poisonMessages, func(i, j int) bool {
		return poisonMessages[i].Message.TimeSent.After(poisonMessages[j].Message.TimeSent)
	})

	// sort.Slice(schemaMessages, func(i, j int) bool {
	// 	return schemaMessages[i].Message.TimeSent.After(schemaMessages[j].Message.TimeSent)
	// })

	if lenPoison > 1000 {
		poisonMessages = poisonMessages[:1000]
	}

	if lenSchema > 1000 {
		schemaMessages = schemaMessages[:1000]
	}
	return poisonMessages, schemaMessages, totalDlsAmount, nil
}

func getDlsMessageById(station models.Station, messageId int, sn StationName, dlsType string) (models.DlsMessageResponsePg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()
	conn, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return models.DlsMessageResponsePg{}, err
	}
	defer conn.Release()
	query := `SELECT * from dls_messages where id=$1 LIMIT 1`
	stmt, err := conn.Conn().Prepare(ctx, "get_dls_msg_by_id", query)
	if err != nil {
		return models.DlsMessageResponsePg{}, err
	}
	rows, err := conn.Conn().Query(ctx, stmt.Name, messageId)
	if err != nil {
		return models.DlsMessageResponsePg{}, err
	}
	defer rows.Close()
	dlsMsgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.DlsMessagePg])
	if err != nil {
		return models.DlsMessageResponsePg{}, err
	}
	if len(dlsMsgs) == 0 {
		return models.DlsMessageResponsePg{}, nil
	}

	poisonedCgs := []models.PoisonedCgPg{}
	var producer models.Producer
	// var dlsMsg models.DlsMessage
	var clientAddress string
	var connectionId string

	msgDetails := models.MessagePayloadDlsPg{
		TimeSent: dlsMsgs[0].MessageDetails.TimeSent,
		Size:     dlsMsgs[0].MessageDetails.Size,
		Data:     hex.EncodeToString([]byte(dlsMsgs[0].MessageDetails.Data)),
		Headers:  dlsMsgs[0].MessageDetails.Headers,
	}
	dlsMsg := models.DlsMessagePg{
		ID:             dlsMsgs[0].ID,
		StationId:      dlsMsgs[0].StationId,
		MessageSeq:     dlsMsgs[0].MessageSeq,
		ProducerId:     dlsMsgs[0].ProducerId,
		PoisonedCgs:    dlsMsgs[0].PoisonedCgs,
		MessageDetails: msgDetails,
		UpdatedAt:      dlsMsgs[0].UpdatedAt,
		MessageType:    dlsMsgs[0].MessageType,
	}

	// if msgType == "poison"
	// poisonedCgs := []models.PoisonedCgPg{}

	if station.IsNative {
		connectionIdHeader := dlsMsg.MessageDetails.Headers["$memphis_connectionId"]
		//This check for backward compatability
		if connectionIdHeader == "" {
			connectionIdHeader = dlsMsg.MessageDetails.Headers["connectionId"]
			if connectionIdHeader == "" {
				return models.DlsMessageResponsePg{}, nil
			}
		}
		connectionId = connectionIdHeader
		_, conn, err := db.GetConnectionByID(connectionId)
		if err != nil {
			return models.DlsMessageResponsePg{}, err
		}
		clientAddress = conn.ClientAddress

		exist, prod, err := db.GetProducerByID(dlsMsg.ProducerId)
		if err != nil {
			return models.DlsMessageResponsePg{}, err
		}
		if !exist {
			return models.DlsMessageResponsePg{}, errors.New("Producer " + prod.Name + " does not exist")
		}
		producer = prod

		// if dlsType == "poison" {
		// cgInfo, err := serv.GetCgInfo(sn, dlsMsg.PoisonedCg.CgName)
		// if err != nil {
		// 	return models.DlsMessageResponse{}, err
		// }

		pc := models.PoisonedCgPg{}
		pCg := dlsMsg.PoisonedCgs
		// if dlsType == "poison" {
		for _, v := range pCg {
			cgInfo, err := serv.GetCgInfo(sn, v)
			if err != nil {
				return models.DlsMessageResponsePg{}, err
			}
			cgMembers, err := GetConsumerGroupMembers(v, station)
			if err != nil {
				return models.DlsMessageResponsePg{}, err
			}
			pc.IsActive, pc.IsDeleted = getCgStatus(cgMembers)

			pc.CgName = v
			pc.TotalPoisonMessages = -1
			pc.MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
			pc.MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
			pc.CgMembers = cgMembers
			pc.UnprocessedMessages = int(cgInfo.NumPending)
			pc.InProcessMessages = cgInfo.NumAckPending
			poisonedCgs = append(poisonedCgs, pc)

		}

		// if dlsType == "schema" {
		// size := len(msg.Subject) + len(dlsMsg.Message.Data) + len(dlsMsg.Message.Headers)
		// dlsMsg.Message.Size = size
		// if dlsMsg.CreatedAt.IsZero() {
		// 	dlsMsg.Message.TimeSent = time.Unix(0, dlsMsg.CreationUnix*1000000)
		// } else {
		// 	dlsMsg.Message.TimeSent = dlsMsg.CreatedAt
		// }
		// }

		for header := range dlsMsg.MessageDetails.Headers {
			if strings.HasPrefix(header, "$memphis") {
				delete(dlsMsg.MessageDetails.Headers, header)
			}
		}
	}

	sort.Slice(poisonedCgs, func(i, j int) bool {
		return poisonedCgs[i].CgName < poisonedCgs[j].CgName
	})

	schemaType := ""
	if station.SchemaName != "" {
		exist, schema, err := db.GetSchemaByName(station.SchemaName)
		if err != nil {
			return models.DlsMessageResponsePg{}, err
		}
		if exist {
			schemaType = schema.Type
		}
	}

	result := models.DlsMessageResponsePg{
		ID:          dlsMsg.ID,
		StationName: station.Name,
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
		Message:     dlsMsg.MessageDetails,
		UpdatedAt:   dlsMsg.UpdatedAt,
		PoisonedCgs: poisonedCgs,
		// ValidationError: dlsMsg.ValidationError,
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

func GetPoisonedCgsByMessage(station models.Station, messageSeq int) ([]models.PoisonedCg, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), db.DbOperationTimeout*time.Second)
	defer cancelfunc()

	connection, err := db.MetadataDbClient.Client.Acquire(ctx)
	if err != nil {
		return []models.PoisonedCg{}, err
	}
	defer connection.Release()

	query := `SELECT dls.poisoned_cgs FROM dls_messages as dls WHERE station_id = $1 AND message_seq = $2 LIMIT 1`

	stmt, err := connection.Conn().Prepare(ctx, "get_dls_messages_by_station_id_and_message_seq", query)
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	rows, err := connection.Conn().Query(ctx, stmt.Name, station.ID, messageSeq)
	if err != nil {
		return []models.PoisonedCg{}, err
	}
	defer rows.Close()

	cgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.PoisonedCgResponseCg])
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	if len(cgs) == 0 {
		return []models.PoisonedCg{}, nil
	}

	poisonedCg := models.PoisonedCg{}
	poisonedCgs := []models.PoisonedCg{}
	for _, cg := range cgs[0].CgName {
		stationName, err := StationNameFromStr(station.Name)
		if err != nil {
			return []models.PoisonedCg{}, err
		}
		cgInfo, err := serv.GetCgInfo(stationName, cg)
		if err != nil {
			return []models.PoisonedCg{}, err
		}
		cgMembers, err := GetConsumerGroupMembers(cg, station)
		if err != nil {
			return []models.PoisonedCg{}, err
		}
		poisonedCg.IsActive, poisonedCg.IsDeleted = getCgStatus(cgMembers)

		poisonedCg.CgName = cg
		poisonedCg.TotalPoisonMessages = -1
		poisonedCg.MaxAckTimeMs = cgMembers[0].MaxAckTimeMs
		poisonedCg.MaxMsgDeliveries = cgMembers[0].MaxMsgDeliveries
		poisonedCg.CgMembers = cgMembers
		poisonedCg.UnprocessedMessages = int(cgInfo.NumPending)
		poisonedCg.InProcessMessages = cgInfo.NumAckPending
		poisonedCgs = append(poisonedCgs, poisonedCg)
	}

	sort.Slice(poisonedCgs, func(i, j int) bool {
		return poisonedCgs[i].CgName < poisonedCgs[j].CgName
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
