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
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
)

const (
	PoisonMessageTitle = "Poison message"
	dlsMsgSep          = "~"
)

type PoisonMessagesHandler struct{ S *Server }

func (s *Server) handleNewUnackedMsg(msg []byte) error {
	var message JSConsumerDeliveryExceededAdvisory
	err := json.Unmarshal(msg, &message)
	if err != nil {
		serv.Errorf("handleNewUnackedMsg at Unmarshal: Error while getting notified about a poison message: %v", err.Error())
		return err
	}

	var streamName string
	var partitionNumber int
	if strings.Contains(message.Stream, "$") {
		streamName = strings.Split(message.Stream, "$")[0]
		partitionNumber, err = strconv.Atoi(strings.Split(message.Stream, "$")[1])
		if err != nil {
			serv.Errorf("handleNewUnackedMsg: Error while converting partition to int: %v", err.Error())
			return err
		}

	} else {
		streamName = message.Stream
		partitionNumber = -1
	}

	accountName := message.Account
	headers := message.Headers
	data := message.Data
	timeSent := message.TimeSent
	var timeSentTimeStamp time.Time
	if timeSent != 0 {
		timeSentTimeStamp = time.Unix(0, timeSent).UTC()
	}
	lenPayload := len(data) + len(headers)
	// backward compatibility
	if accountName == _EMPTY_ {
		accountName = s.MemphisGlobalAccountString()
	}
	stationName := StationNameFromStreamName(streamName)
	_, station, err := db.GetStationByName(stationName.Ext(), accountName)
	if err != nil {
		serv.Errorf("[tenant: %v]handleNewUnackedMsg at GetStationByName: station: %v, Error while getting notified about a poison message: %v", accountName, stationName.Ext(), err.Error())
		return err
	}
	if !station.DlsConfigurationPoison {
		return nil
	}

	cgName := message.Consumer
	cgName = revertDelimiters(cgName)
	messageSeq := message.StreamSeq
	// backward compatibility
	if data == nil {
		poisonMessageContent, err := s.memphisGetMessage(accountName, message.Stream, uint64(messageSeq))
		if err != nil {
			if IsNatsErr(err, JSNoMessageFoundErr) {
				return nil
			}
			serv.Errorf("[tenant: %v]handleNewUnackedMsg at memphisGetMessage: station: %v, Error while getting notified about a poison message: %v", accountName, stationName.Ext(), err.Error())
			return err
		}

		timeSentTimeStamp = poisonMessageContent.Time
		data = poisonMessageContent.Data
		headers = poisonMessageContent.Header
		lenPayload = len(poisonMessageContent.Data) + len(poisonMessageContent.Header)
	}
	var headersJson map[string]string
	if headers != nil {
		headersJson, err = DecodeHeader(headers)
		if err != nil {
			serv.Errorf("handleNewUnackedMsg: %v", err.Error())
			return err
		}
	}

	producedByHeader := _EMPTY_
	poisonedCgs := []string{}
	if station.IsNative {
		producedByHeader = headersJson["$memphis_producedBy"]
		if producedByHeader == _EMPTY_ {
			producedByHeader = "unknown"
		}

		if producedByHeader == "$memphis_dls" { // skip poison messages which have been resent
			return nil
		}
		poisonedCgs = append(poisonedCgs, cgName)
	}

	messageDetails := models.MessagePayload{
		TimeSent: timeSentTimeStamp,
		Size:     lenPayload,
		Data:     hex.EncodeToString(data),
		Headers:  headersJson,
	}

	dlsMsgId, updated, err := db.StorePoisonMsg(station.ID, int(messageSeq), cgName, producedByHeader, poisonedCgs, messageDetails, station.TenantName, partitionNumber, "")
	if err != nil {
		serv.Errorf("[tenant: %v]handleNewUnackedMsg at StorePoisonMsg: Error while getting notified about a poison message: %v", station.TenantName, err.Error())
		return err
	}
	if !updated {
		err = s.sendToDlsStation(station, data, headersJson, "unacked", _EMPTY_)
		if err != nil {
			serv.Errorf("[tenant: %v]handleNewUnackedMsg at sendToDlsStation: station: %v, Error while getting notified about a poison message: %v", station.TenantName, station.DlsStation, err.Error())
			return err
		}
	}

	if dlsMsgId == 0 { // nothing to do
		return nil
	}

	idForUrl := strconv.Itoa(dlsMsgId)
	var msgUrl = s.opts.UiHost + "/stations/" + stationName.Ext() + "/" + idForUrl
	err = s.SendNotification(station.TenantName, PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, PoisonMAlert)
	if err != nil {
		serv.Warnf("[tenant: %v]handleNewUnackedMsg at SendNotification: Error while sending a poison message notification: %v", station.TenantName, err.Error())
		return nil
	}
	return nil
}

func (s *Server) handleSchemaverseDlsMsg(msg []byte) error {
	tenantName, stringMessage, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("handleSchemaverseDlsMsg at getTenantNameAndMessage: %v", err.Error())
		return err
	}
	var message models.SchemaVerseDlsMessageSdk
	err = json.Unmarshal([]byte(stringMessage), &message)
	if err != nil {
		serv.Errorf("[tenant: %v]handleSchemaverseDlsMsg: %v", tenantName, err.Error())
		return err
	}

	stationName := StationNameFromStreamName(message.StationName)
	exist, station, err := db.GetStationByName(stationName.Ext(), tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]handleSchemaverseDlsMsg: %v", tenantName, err.Error())
		return err
	}
	if !exist {
		serv.Warnf("[tenant: %v]handleSchemaverseDlsMsg: station %v couldn't been found", tenantName, stationName.Ext())
		return nil
	}

	message.Message.TimeSent = time.Now()
	_, err = db.InsertSchemaverseDlsMsg(station.ID, 0, message.Producer.Name, []string{}, models.MessagePayload(message.Message), message.ValidationError, tenantName, message.PartitionNumber)
	if err != nil {
		serv.Errorf("[tenant: %v]handleSchemaverseDlsMsg: %v", tenantName, err.Error())
		return err
	}
	data, err := hex.DecodeString(message.Message.Data)
	if err != nil {
		serv.Errorf("[tenant: %v]handleSchemaverseDlsMsg at DecodeString: %v", tenantName, err.Error())
		return err
	}
	err = s.sendToDlsStation(station, data, message.Message.Headers, "failed_schema", _EMPTY_)
	if err != nil {
		serv.Errorf("[tenant: %v]handleSchemaverseDlsMsg at sendToDlsStation: station: %v, Error while getting notified about a poison message: %v", tenantName, station.DlsStation, err.Error())
		return err
	}

	return nil
}

func (s *Server) handleNackedDlsMsg(msg []byte) error {
	tenantName, stringMessage, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("handleNackedDlsMsg at getTenantNameAndMessage: %v", err.Error())
		return err
	}
	var message models.NackedDlsMessageSdk
	err = json.Unmarshal([]byte(stringMessage), &message)
	if err != nil {
		serv.Errorf("[tenant: %v]handleNackedDlsMsg: %v", tenantName, err.Error())
		return err
	}

	if message.Partition == 0 {
		serv.Errorf("[tenant: %v]handleNackedDlsMsg - missing partition number: %v", tenantName, err.Error())
		return err
	}

	stationName := StationNameFromStreamName(message.StationName)
	streamName := stationName.Intern() + "$" + strconv.Itoa(message.Partition)
	exist, station, err := db.GetStationByName(stationName.Ext(), tenantName)
	if err != nil {
		serv.Errorf("[tenant: %v]handleNackedDlsMsg: %v", tenantName, err.Error())
		return err
	}
	if !exist {
		serv.Warnf("[tenant: %v]handleNackedDlsMsg: station %v couldn't been found", tenantName, stationName.Ext())
		return nil
	}
	if !station.DlsConfigurationPoison {
		return nil
	}

	poisonMessageContent, err := s.memphisGetMessage(tenantName, streamName, uint64(message.Seq))
	if err != nil {
		if IsNatsErr(err, JSNoMessageFoundErr) {
			return nil
		}
		serv.Errorf("[tenant: %v]handleNackedDlsMsg at memphisGetMessage: station: %v, Error while getting notified about a poison message: %v", tenantName, stationName.Ext(), err.Error())
		return err
	}

	
	timeSentTimeStamp := poisonMessageContent.Time
	data := poisonMessageContent.Data
	lenPayload := len(poisonMessageContent.Data) + len(poisonMessageContent.Header)
	headers := poisonMessageContent.Header
	var headersJson map[string]string
	if headers != nil {
		headersJson, err = DecodeHeader(headers)
		if err != nil {
			serv.Errorf("handleNackedDlsMsg: %v", err.Error())
			return err
		}
	}

	producedByHeader := _EMPTY_
	poisonedCgs := []string{}
	producedByHeader = headersJson["$memphis_producedBy"]
	if producedByHeader == _EMPTY_ {
		producedByHeader = "unknown"
	}
	poisonedCgs = append(poisonedCgs, message.CgName)

	messageDetails := models.MessagePayload{
		TimeSent: timeSentTimeStamp,
		Size:     lenPayload,
		Data:     hex.EncodeToString(data),
		Headers:  headersJson,
	}

	dlsMsgId, updated, err := db.StorePoisonMsg(station.ID, int(message.Seq), message.CgName, producedByHeader, poisonedCgs, messageDetails, tenantName, message.Partition, message.Error)
	if err != nil {
		serv.Errorf("[tenant: %v]handleNackedDlsMsg at StorePoisonMsg: Error while getting notified about a poison message: %v", station.TenantName, err.Error())
		return err
	}
	if !updated {
		err = s.sendToDlsStation(station, data, headersJson, "unacked", _EMPTY_)
		if err != nil {
			serv.Errorf("[tenant: %v]handleNackedDlsMsg at sendToDlsStation: station: %v, Error while getting notified about a poison message: %v", station.TenantName, station.DlsStation, err.Error())
			return err
		}
	}

	if dlsMsgId == 0 { // nothing to do
		return nil
	}

	idForUrl := strconv.Itoa(dlsMsgId)
	var msgUrl = s.opts.UiHost + "/stations/" + stationName.Ext() + "/" + idForUrl
	err = s.SendNotification(station.TenantName, PoisonMessageTitle, "Poison message has been identified, for more details head to: "+msgUrl, PoisonMAlert)
	if err != nil {
		serv.Warnf("[tenant: %v]handleNackedDlsMsg at SendNotification: Error while sending a poison message notification: %v", station.TenantName, err.Error())
		return nil
	}

	return nil
}

func (pmh PoisonMessagesHandler) GetDlsMsgsByStationLight(station models.Station, partitionNumber int) ([]models.LightDlsMessageResponse, []models.LightDlsMessageResponse, []models.LightDlsMessageResponse, int, error) {
	poisonMessages := make([]models.LightDlsMessageResponse, 0)
	schemaMessages := make([]models.LightDlsMessageResponse, 0)
	functionsMessages := make([]models.LightDlsMessageResponse, 0)

	var dlsMsgs []models.DlsMessage
	var err error
	if partitionNumber == -1 {
		dlsMsgs, err = db.GetDlsMsgsByStationId(station.ID)
		if err != nil {
			return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, err
		}
	} else {
		dlsMsgs, err = db.GetDlsMsgsByStationAndPartition(station.ID, partitionNumber)
		if err != nil {
			return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, err
		}
	}

	for _, v := range dlsMsgs {
		data := v.MessageDetails.Data
		if len(data) > 80 { // get the first chars for preview needs
			data = data[0:80]
		}
		messageDetails := models.MessagePayload{
			TimeSent: v.MessageDetails.TimeSent,
			Size:     v.MessageDetails.Size,
			Data:     data,
			Headers:  v.MessageDetails.Headers,
		}
		switch v.MessageType {
		case "poison":
			poisonMessages = append(poisonMessages, models.LightDlsMessageResponse{MessageSeq: v.MessageSeq, ID: v.ID, Message: messageDetails})
		case "schema":
			messageDetails.Size = len(v.MessageDetails.Data) + len(v.MessageDetails.Headers)
			schemaMessages = append(schemaMessages, models.LightDlsMessageResponse{MessageSeq: v.MessageSeq, ID: v.ID, Message: messageDetails})
		case "functions":
			functionsMessages = append(functionsMessages, models.LightDlsMessageResponse{MessageSeq: v.MessageSeq, ID: v.ID, Message: messageDetails})
		}
	}

	lenPoison, lenSchema, lenFunctions := len(poisonMessages), len(schemaMessages), len(functionsMessages)
	totalDlsAmount := 0
	if len(dlsMsgs) >= 0 {
		totalDlsAmount, err = db.CountDlsMsgsByStationAndPartition(station.ID, partitionNumber)
		if err != nil {
			return []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, []models.LightDlsMessageResponse{}, 0, err
		}
	}

	sort.Slice(poisonMessages, func(i, j int) bool {
		return poisonMessages[i].Message.TimeSent.After(poisonMessages[j].Message.TimeSent)
	})

	sort.Slice(schemaMessages, func(i, j int) bool {
		return schemaMessages[i].Message.TimeSent.After(schemaMessages[j].Message.TimeSent)
	})

	sort.Slice(functionsMessages, func(i, j int) bool {
		return functionsMessages[i].Message.TimeSent.After(functionsMessages[j].Message.TimeSent)
	})

	if lenPoison > 1000 {
		poisonMessages = poisonMessages[:1000]
	}

	if lenSchema > 1000 {
		schemaMessages = schemaMessages[:1000]
	}

	if lenFunctions > 1000 {
		functionsMessages = functionsMessages[:1000]
	}
	return poisonMessages, schemaMessages, functionsMessages, totalDlsAmount, nil
}

func (pmh PoisonMessagesHandler) GetDlsMessageDetailsById(messageId int, dlsType string, tenantName string) (models.DlsMessageResponse, error) {
	exist, dlsMessage, err := db.GetDlsMessageById(messageId)
	if err != nil {
		return models.DlsMessageResponse{}, err
	}
	if !exist {
		return models.DlsMessageResponse{}, errors.New("dls message does not exists")
	}
	exist, station, err := db.GetStationById(dlsMessage.StationId, dlsMessage.TenantName)
	if err != nil {
		return models.DlsMessageResponse{}, err
	}
	if !exist {
		return models.DlsMessageResponse{}, fmt.Errorf("Station %v does not exists", station.Name)
	}

	sn, err := StationNameFromStr(station.Name)
	if err != nil {
		return models.DlsMessageResponse{}, err
	}

	poisonedCgs := []models.PoisonedCg{}
	isActive := false

	msgDetails := models.MessagePayload{
		TimeSent: dlsMessage.MessageDetails.TimeSent,
		Size:     dlsMessage.MessageDetails.Size,
		Data:     dlsMessage.MessageDetails.Data,
		Headers:  dlsMessage.MessageDetails.Headers,
	}
	dlsMsg := models.DlsMessage{
		ID:              dlsMessage.ID,
		StationId:       dlsMessage.StationId,
		PartitionNumber: dlsMessage.PartitionNumber,
		MessageSeq:      dlsMessage.MessageSeq,
		ProducerName:    dlsMessage.ProducerName,
		PoisonedCgs:     dlsMessage.PoisonedCgs,
		MessageDetails:  msgDetails,
		UpdatedAt:       dlsMessage.UpdatedAt,
		MessageType:     dlsMessage.MessageType,
		ValidationError: dlsMessage.ValidationError,
	}

	if station.IsNative {
		exist, prod, err := db.GetProducerByNameAndStationID(dlsMsg.ProducerName, dlsMsg.StationId)
		if err != nil {
			return models.DlsMessageResponse{}, err
		}
		if exist {
			isActive = prod.IsActive
		}

		pc := models.PoisonedCg{}
		pCg := dlsMsg.PoisonedCgs
		for _, v := range pCg {
			cgMembers, err := GetConsumerGroupMembers(v, station)
			if err != nil {
				serv.Errorf("[tenant: %v]GetDlsMessageDetailsById at GetConsumerGroupMembers: %v", station.TenantName, err.Error())
				return models.DlsMessageResponse{}, err
			}
			cgInfo, err := serv.GetCgInfo(station.TenantName, sn, v, cgMembers[0].PartitionsList)
			if err != nil {
				serv.Errorf("[tenant: %v]GetDlsMessageDetailsById at GetCgInfo: %v", station.TenantName, err.Error())
				return models.DlsMessageResponse{}, err
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

		if dlsType == "schema" {
			size := len(dlsMessage.MessageDetails.Data) + len(dlsMessage.MessageDetails.Headers)
			dlsMsg.MessageDetails.Size = size
		}

		for header := range dlsMsg.MessageDetails.Headers {
			if strings.HasPrefix(header, MEMPHIS_GLOBAL_ACCOUNT) {
				delete(dlsMsg.MessageDetails.Headers, header)
			}
		}
	}

	sort.Slice(poisonedCgs, func(i, j int) bool {
		return poisonedCgs[i].CgName < poisonedCgs[j].CgName
	})

	schemaType := _EMPTY_
	if station.SchemaName != _EMPTY_ {
		exist, schema, err := db.GetSchemaByName(station.SchemaName, station.TenantName)
		if err != nil {
			return models.DlsMessageResponse{}, err
		}
		if exist {
			schemaType = schema.Type
		}
	}

	result := models.DlsMessageResponse{
		ID:          dlsMsg.ID,
		StationName: station.Name,
		SchemaType:  schemaType,
		MessageSeq:  dlsMsg.MessageSeq,
		Producer: models.ProducerDetailsResp{
			Name:     dlsMsg.ProducerName,
			IsActive: isActive,
		},
		Message:         dlsMsg.MessageDetails,
		UpdatedAt:       dlsMsg.UpdatedAt,
		PoisonedCgs:     poisonedCgs,
		ValidationError: dlsMsg.ValidationError,
	}

	return result, nil
}

func GetPoisonedCgsByMessage(station models.Station, messageSeq, partitionNumber int) ([]models.PoisonedCg, error) {
	var dlsMsg models.DlsMessage
	_, dlsMsg, err := db.GetMsgByStationIdAndMsgSeq(station.ID, messageSeq, partitionNumber)
	if err != nil {
		return []models.PoisonedCg{}, err
	}

	cgs := dlsMsg.PoisonedCgs

	poisonedCg := models.PoisonedCg{}
	poisonedCgs := []models.PoisonedCg{}
	for _, cg := range cgs {
		stationName, err := StationNameFromStr(station.Name)
		if err != nil {
			return []models.PoisonedCg{}, err
		}
		cgMembers, err := GetConsumerGroupMembers(cg, station)
		if err != nil {
			serv.Errorf("[tenant: %v]GetPoisonedCgsByMessage at GetConsumerGroupMembers: %v", station.TenantName, err.Error())
			return []models.PoisonedCg{}, err
		}

		cgInfo, err := serv.GetCgInfo(station.TenantName, stationName, cg, cgMembers[0].PartitionsList)
		if err != nil {
			serv.Errorf("[tenant: %v]GetPoisonedCgsByMessage at GetCgInfo: %v", station.TenantName, err.Error())
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

func (s *Server) sendToDlsStation(station models.Station, messagePayload []byte, headers map[string]string, dlsType, functionName string) error {
	if station.DlsStation != _EMPTY_ {
		exist, dlsStation, err := db.GetStationByName(station.DlsStation, station.TenantName)
		if err != nil {
			return err
		}
		if exist {
			dlsStationName, err := StationNameFromStr(dlsStation.Name)
			if err != nil {
				return err
			}
			subject := _EMPTY_
			shouldRoundRobin := false
			if dlsStation.Version == 0 {
				subject = fmt.Sprintf("%s.final", dlsStationName.Intern())
			} else {
				shouldRoundRobin = true
			}

			if shouldRoundRobin {
				rand.Seed(time.Now().UnixNano())
				randomIndex := 0
				if len(dlsStation.PartitionsList) > 1 {
					randomIndex = rand.Intn(len(dlsStation.PartitionsList) - 1)
				}
				subject = fmt.Sprintf("%s$%v.final", dlsStationName.Intern(), dlsStation.PartitionsList[randomIndex])
			}

			acc, err := s.LookupAccount(station.TenantName)
			if err != nil {
				return err
			}
			headers["station"] = station.Name
			headers["type"] = dlsType
			if dlsType == "functions" {
				headers["function_name"] = functionName
			}
			s.sendInternalAccountMsgWithHeadersWithEcho(acc, subject, messagePayload, headers)
		}
	}
	return nil
}
