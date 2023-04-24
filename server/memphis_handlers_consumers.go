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
	"errors"
	"sort"
	"strconv"

	"memphis/analytics"
	"memphis/db"
	"memphis/models"
	"memphis/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/utils/strings/slices"
)

type ConsumersHandler struct{ S *Server }

const (
	consumerObjectName = "Consumer"
)

func validateConsumerName(consumerName string) error {
	return validateName(consumerName, consumerObjectName)
}

func validateConsumerType(consumerType string) error {
	if consumerType != "application" && consumerType != "connector" {
		return errors.New("consumer type has to be one of the following application/connector")
	}
	return nil
}

func isConsumerGroupExist(consumerGroup string, stationId int) (bool, models.Consumer, error) {
	exist, consumer, err := db.GetActiveConsumerByCG(consumerGroup, stationId)
	if err != nil {
		return false, models.Consumer{}, err
	} else if !exist {
		return false, models.Consumer{}, nil
	}
	return true, consumer, nil
}

func GetConsumerGroupMembers(cgName string, station models.Station) ([]models.CgMember, error) {
	consumers, err := db.GetConsumerGroupMembers(cgName, station.ID)
	if err != nil {
		return consumers, err
	}

	var dedupedConsumers []models.CgMember
	consumersNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersNames, consumer.Name) {
			continue
		}
		consumersNames = append(consumersNames, consumer.Name)
		dedupedConsumers = append(dedupedConsumers, consumer)
	}

	return dedupedConsumers, nil
}

func (s *Server) createConsumerDirectV0(c *client, reply string, ccr createConsumerRequestV0, requestVersion int) {
	err := s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, requestVersion, 1, -1)
	respondWithErr(s, reply, err)
}

func (s *Server) createConsumerDirectCommon(c *client, consumerName, cStationName, cGroup, cType, connectionId string, maxAckTime, maxMsgDeliveries, requestVersion int, startConsumeFromSequence uint64, lastMessages int64) error {
	name := strings.ToLower(consumerName)
	err := validateConsumerName(name)
	if err != nil {
		serv.Warnf("createConsumerDirectCommon: Failed creating consumer " + consumerName + " at station " + cStationName + ": " + err.Error())
		return err
	}

	consumerGroup := strings.ToLower(cGroup)
	if consumerGroup != "" {
		err = validateConsumerName(consumerGroup)
		if err != nil {
			serv.Warnf("createConsumerDirectCommon: Failed creating consumer " + consumerName + " at station " + cStationName + ": " + err.Error())
			return err
		}
	} else {
		consumerGroup = name
	}

	consumerType := strings.ToLower(cType)
	err = validateConsumerType(consumerType)
	if err != nil {
		serv.Warnf("createConsumerDirectCommon: Failed creating consumer " + consumerName + " at station " + cStationName + ": " + err.Error())
		return err
	}

	exist, connection, err := db.GetConnectionByID(connectionId)
	if err != nil {
		errMsg := "Consumer " + consumerName + ": " + err.Error()
		serv.Errorf("createConsumerDirectCommon: " + errMsg)
		return err
	}
	if !exist {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": Connection ID " + connectionId + " was not found"
		serv.Warnf("createConsumerDirectCommon: " + errMsg)
		return errors.New(errMsg)
	}
	if !connection.IsActive {
		serv.Warnf("createConsumerDirectCommon: Failed creating consumer " + consumerName + " at station " + cStationName + ": Connection is not active")
		return errors.New("connection is not active")
	}

	stationName, err := StationNameFromStr(cStationName)
	if err != nil {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
		serv.Warnf("createConsumerDirectCommon: " + errMsg)
		return err
	}

	exist, user, err := db.GetUserByUserId(connection.CreatedBy)
	if err != nil {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
		serv.Errorf("createConsumerDirectCommon: " + errMsg)
		return err
	}
	if !exist {
		serv.Warnf("createProducerDirectCommon: user %v is not exists", connection.CreatedBy)
		return err
	}

	exist, station, err := db.GetStationByName(stationName.Ext(), user.TenantName)
	if err != nil {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
		serv.Errorf("createConsumerDirectCommon: " + errMsg)
		return err
	}

	if !exist {
		var created bool
		station, created, err = CreateDefaultStation(s, stationName, connection.CreatedBy, user.Username)
		if err != nil {
			errMsg := "creating default station error: Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
			serv.Warnf("createConsumerDirectCommon: " + errMsg)
			return err
		}

		if created {
			message := "Station " + stationName.Ext() + " has been created by user " + user.Username
			serv.Noticef(message)
			var auditLogs []interface{}
			newAuditLog := models.AuditLog{
				StationName:       stationName.Ext(),
				Message:           message,
				CreatedBy:         connection.CreatedBy,
				CreatedByUsername: connection.CreatedByUsername,
				CreatedAt:         time.Now(),
			}
			auditLogs = append(auditLogs, newAuditLog)
			err = CreateAuditLogs(auditLogs)
			if err != nil {
				errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
				serv.Errorf("createConsumerDirect: " + errMsg)
			}

			shouldSendAnalytics, _ := shouldSendAnalytics()
			if shouldSendAnalytics {
				param := analytics.EventParam{
					Name:  "station-name",
					Value: stationName.Ext(),
				}
				analyticsParams := []analytics.EventParam{param}
				analytics.SendEventWithParams(strconv.Itoa(connection.CreatedBy), analyticsParams, "user-create-station-sdk")
			}
		}
	}

	consumerGroupExist, consumerFromGroup, err := isConsumerGroupExist(consumerGroup, station.ID)
	if err != nil {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
		serv.Errorf("createConsumerDirectCommon: " + errMsg)
		return err
	}

	exist, newConsumer, rowsUpdated, err := db.InsertNewConsumer(name, station.ID, consumerType, connectionId, connection.CreatedBy, user.Username, consumerGroup, maxAckTime, maxMsgDeliveries, startConsumeFromSequence, lastMessages)
	if err != nil {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
		serv.Errorf("createConsumerDirectCommon: " + errMsg)
		return err
	}
	if exist {
		errMsg := "Consumer " + consumerName + " at station " + cStationName + ": Consumer name has to be unique per station"
		serv.Warnf("createConsumerDirectCommon: " + errMsg)
		return errors.New("memphis: " + errMsg)
	}

	if rowsUpdated == 1 {
		message := "Consumer " + name + " has been created by user " + user.Username
		serv.Noticef(message)
		if consumerGroupExist {
			if requestVersion == 1 {
				if newConsumer.StartConsumeFromSeq != consumerFromGroup.StartConsumeFromSeq || newConsumer.LastMessages != consumerFromGroup.LastMessages {
					errMsg := errors.New("consumer already exists with different uneditable configuration parameters (StartConsumeFromSequence/LastMessages)")
					serv.Warnf("createConsumerDirectCommon: " + errMsg.Error())
					return errMsg
				}
			}

			if newConsumer.MaxAckTimeMs != consumerFromGroup.MaxAckTimeMs || newConsumer.MaxMsgDeliveries != consumerFromGroup.MaxMsgDeliveries {
				err := s.CreateConsumer(newConsumer, station)
				if err != nil {
					if IsNatsErr(err, JSStreamNotFoundErr) {
						errMsg := "Consumer " + consumerName + " at station " + cStationName + ": station does not exist"
						serv.Warnf("createConsumerDirectCommon: " + errMsg)
					} else {
						errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
						serv.Errorf("createConsumerDirectCommon: " + errMsg)
					}
					return err
				}
			}
		} else {
			err := s.CreateConsumer(newConsumer, station)
			if err != nil {
				if IsNatsErr(err, JSStreamNotFoundErr) {
					errMsg := "Consumer " + consumerName + " at station " + cStationName + ": station does not exist"
					serv.Warnf("createConsumerDirectCommon: " + errMsg)
				} else {
					errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
					serv.Errorf("createConsumerDirectCommon: " + errMsg)
				}
				return err
			}
		}
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Ext(),
			Message:           message,
			CreatedBy:         connection.CreatedBy,
			CreatedByUsername: connection.CreatedByUsername,
			CreatedAt:         time.Now(),
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			errMsg := "Consumer " + consumerName + " at station " + cStationName + ": " + err.Error()
			serv.Errorf("createConsumerDirectCommon: " + errMsg)
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			param := analytics.EventParam{
				Name:  "consumer-name",
				Value: newConsumer.Name,
			}
			analyticsParams := []analytics.EventParam{param}
			analytics.SendEventWithParams(strconv.Itoa(connection.CreatedBy), analyticsParams, "user-create-consumer-sdk")
		}
	}
	return nil
}

func (s *Server) createConsumerDirect(c *client, reply string, msg []byte) {
	var ccr createConsumerRequestV1
	var resp createConsumerResponse
	if err := json.Unmarshal(msg, &ccr); err != nil || ccr.RequestVersion < 1 {
		var ccrV0 createConsumerRequestV0
		if err := json.Unmarshal(msg, &ccrV0); err != nil {
			s.Errorf("createConsumerDirect: Failed creating consumer: %v\n%v", err.Error(), string(msg))
			respondWithRespErr(s, reply, err, &resp)
			return
		}
		s.createConsumerDirectV0(c, reply, ccrV0, ccr.RequestVersion)
		return
	}

	if ccr.StartConsumeFromSequence <= 0 {
		errMsg := errors.New("startConsumeFromSequence has to be a positive number")
		serv.Warnf("createConsumerDirect: " + errMsg.Error())
		respondWithErr(s, reply, errMsg)
		return
	}

	if ccr.LastMessages < -1 {
		errMsg := errors.New("min value for LastMessages is -1")
		serv.Warnf("createConsumerDirect: " + errMsg.Error())
		respondWithErr(s, reply, errMsg)
		return
	}

	if ccr.StartConsumeFromSequence > 1 && ccr.LastMessages > -1 {
		errMsg := errors.New("consumer creation options can't contain both startConsumeFromSequence and lastMessages")
		serv.Warnf("createConsumerDirect: " + errMsg.Error())
		respondWithErr(s, reply, errMsg)
		return
	}

	err := s.createConsumerDirectCommon(c, ccr.Name, ccr.StationName, ccr.ConsumerGroup, ccr.ConsumerType, ccr.ConnectionId, ccr.MaxAckTimeMillis, ccr.MaxMsgDeliveries, 1, ccr.StartConsumeFromSequence, ccr.LastMessages)
	respondWithErr(s, reply, err)
}

func (ch ConsumersHandler) GetAllConsumers(c *gin.Context) {
	consumers, err := db.GetAllConsumers()
	if err != nil {
		serv.Errorf("GetAllConsumers: " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (ch ConsumersHandler) GetCgsByStation(stationName StationName, station models.Station) ([]models.Cg, []models.Cg, []models.Cg, error) { // for socket io endpoint
	var cgs []models.Cg
	consumers, err := db.GetAllConsumersByStation(station.ID)
	if err != nil {
		return cgs, cgs, cgs, err
	}

	if len(consumers) == 0 {
		return []models.Cg{}, []models.Cg{}, []models.Cg{}, nil
	}

	m := make(map[string]*models.Cg)
	consumersGroupNames := []string{}

	for _, consumer := range consumers {
		if slices.Contains(consumersGroupNames, consumer.ConsumersGroup+consumer.Name) {
			continue
		}
		consumersGroupNames = append(consumersGroupNames, consumer.ConsumersGroup+consumer.Name)

		var cg *models.Cg
		if m[consumer.ConsumersGroup] == nil {
			cg = &models.Cg{
				Name:                  consumer.ConsumersGroup,
				MaxAckTimeMs:          consumer.MaxAckTimeMs,
				MaxMsgDeliveries:      consumer.MaxMsgDeliveries,
				ConnectedConsumers:    []models.ExtendedConsumer{},
				DisconnectedConsumers: []models.ExtendedConsumer{},
				DeletedConsumers:      []models.ExtendedConsumer{},
				IsActive:              consumer.IsActive,
				IsDeleted:             consumer.IsDeleted,
				LastStatusChangeDate:  consumer.CreatedAt,
			}
			m[consumer.ConsumersGroup] = cg
		} else {
			cg = m[consumer.ConsumersGroup]
		}

		consumerRes := models.ExtendedConsumer{
			ID:                consumer.ID,
			Name:              consumer.Name,
			CreatedByUsername: consumer.CreatedByUsername,
			CreatedAt:         consumer.CreatedAt,
			IsActive:          consumer.IsActive,
			ClientAddress:     consumer.ClientAddress,
			ConsumersGroup:    consumer.ConsumersGroup,
			MaxAckTimeMs:      consumer.MaxAckTimeMs,
			MaxMsgDeliveries:  consumer.MaxMsgDeliveries,
			StationName:       consumer.StationName,
		}

		if consumer.IsActive {
			cg.ConnectedConsumers = append(cg.ConnectedConsumers, consumerRes)
		} else if !consumer.IsDeleted && !consumer.IsActive {
			cg.DisconnectedConsumers = append(cg.DisconnectedConsumers, consumerRes)
		} else if consumer.IsDeleted {
			cg.DeletedConsumers = append(cg.DeletedConsumers, consumerRes)
		}
	}

	var connectedCgs []models.Cg
	var disconnectedCgs []models.Cg
	var deletedCgs []models.Cg

	for _, cg := range m {
		if cg.IsDeleted {
			cg.IsActive = false
			cg.IsDeleted = true
		} else { // not deleted
			cgInfo, err := ch.S.GetCgInfo(stationName, cg.Name)
			if err != nil {
				continue // ignoring cases where the consumer exist in memphis but not in nats
			}

			totalPoisonMsgs, err := db.GetTotalPoisonMsgsPerCg(cg.Name, station.ID)
			if err != nil {
				return []models.Cg{}, []models.Cg{}, []models.Cg{}, err
			}

			cg.InProcessMessages = cgInfo.NumAckPending
			cg.UnprocessedMessages = int(cgInfo.NumPending)
			cg.PoisonMessages = totalPoisonMsgs
		}

		if len(cg.ConnectedConsumers) > 0 {
			connectedCgs = append(connectedCgs, *cg)
		} else if len(cg.DisconnectedConsumers) > 0 {
			disconnectedCgs = append(disconnectedCgs, *cg)
		} else {
			deletedCgs = append(deletedCgs, *cg)
		}
	}

	if len(connectedCgs) == 0 {
		connectedCgs = []models.Cg{}
	}

	if len(disconnectedCgs) == 0 {
		disconnectedCgs = []models.Cg{}
	}

	if len(deletedCgs) == 0 {
		deletedCgs = []models.Cg{}
	}

	sort.Slice(connectedCgs, func(i, j int) bool {
		return connectedCgs[j].LastStatusChangeDate.Before(connectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(disconnectedCgs, func(i, j int) bool {
		return disconnectedCgs[j].LastStatusChangeDate.Before(disconnectedCgs[i].LastStatusChangeDate)
	})
	sort.Slice(deletedCgs, func(i, j int) bool {
		return deletedCgs[j].LastStatusChangeDate.Before(deletedCgs[i].LastStatusChangeDate)
	})
	return connectedCgs, disconnectedCgs, deletedCgs, nil
}

// TODO fix it
func (ch ConsumersHandler) GetAllConsumersByStation(c *gin.Context) { // for REST endpoint
	var body models.GetAllConsumersByStationSchema
	ok := utils.Validate(c, &body, false, nil)
	if !ok {
		return
	}

	sn, err := StationNameFromStr(body.StationName)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	user, err := getUserDetailsFromMiddleware(c)
	if err != nil {
		serv.Errorf("GetAllConsumersByStation: " + err.Error())
		c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	exist, station, err := db.GetStationByName(sn.Ext(), user.TenantName)
	if err != nil {
		serv.Errorf("GetAllConsumersByStation: At station " + body.StationName + ": " + err.Error())
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}
	if !exist {
		serv.Warnf("GetAllConsumersByStation: Station " + body.StationName + " does not exist")
		c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Station does not exist"})
		return
	}

	consumers, err := db.GetAllConsumersByStation(station.ID)
	if err != nil {
		errMsg := "Station " + body.StationName + ": " + err.Error()
		serv.Errorf("GetAllConsumersByStation: " + errMsg)
		c.AbortWithStatusJSON(500, gin.H{"message": "Server error"})
		return
	}

	if len(consumers) == 0 {
		c.IndentedJSON(200, []string{})
	} else {
		c.IndentedJSON(200, consumers)
	}
}

func (s *Server) destroyConsumerDirect(c *client, reply string, msg []byte) {
	var dcr destroyConsumerRequest
	if err := json.Unmarshal(msg, &dcr); err != nil {
		s.Errorf("destroyConsumerDirect: %v", err.Error())
		respondWithErr(s, reply, err)
		return
	}

	stationName, err := StationNameFromStr(dcr.StationName)
	if err != nil {
		errMsg := "Station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	name := strings.ToLower(dcr.ConsumerName)
	_, station, err := db.GetStationByName(stationName.Ext(), c.acc.GetName())
	if err != nil {
		errMsg := "Station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}
	exist, consumer, err := db.DeleteConsumer(name, station.ID)
	if !exist {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + " does not exist"
		serv.Warnf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, errors.New(errMsg))
		return
	}
	if err != nil {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	// ensure not part of an active consumer group
	count, err := db.CountActiveConsumersInCG(consumer.ConsumersGroup, station.ID)
	if err != nil {
		errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
		serv.Errorf("DestroyConsumer: " + errMsg)
		respondWithErr(s, reply, err)
		return
	}

	deleted := false
	if count == 0 { // no other members in this group
		err = s.RemoveConsumer(stationName, consumer.ConsumersGroup)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}
		if err == nil {
			deleted = true
		}
		err = db.RemovePoisonedCg(station.ID, consumer.ConsumersGroup)
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}
	}

	if deleted {
		username := c.memphisInfo.username
		if username == "" {
			username = dcr.Username
		}
		_, user, err := db.GetUserByUsername(username, c.acc.GetName())
		if err != nil && !IsNatsErr(err, JSConsumerNotFoundErr) && !IsNatsErr(err, JSStreamNotFoundErr) {
			errMsg := "Consumer group " + consumer.ConsumersGroup + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
			respondWithErr(s, reply, err)
			return
		}
		message := "Consumer " + name + " has been deleted by user " + username
		serv.Noticef(message)
		var auditLogs []interface{}
		newAuditLog := models.AuditLog{
			StationName:       stationName.Ext(),
			Message:           message,
			CreatedBy:         user.ID,
			CreatedByUsername: user.Username,
			CreatedAt:         time.Now(),
		}
		auditLogs = append(auditLogs, newAuditLog)
		err = CreateAuditLogs(auditLogs)
		if err != nil {
			errMsg := "Consumer " + dcr.ConsumerName + " at station " + dcr.StationName + ": " + err.Error()
			serv.Errorf("DestroyConsumer: " + errMsg)
		}

		shouldSendAnalytics, _ := shouldSendAnalytics()
		if shouldSendAnalytics {
			analytics.SendEvent(username, "user-remove-consumer-sdk")
		}
	}

	respondWithErr(s, reply, nil)
}

func (ch ConsumersHandler) ReliveConsumers(connectionId string) error {
	err := db.UpdateConsumersConnection(connectionId, false)
	if err != nil {
		serv.Errorf("ReliveConsumers: " + err.Error())
		return err
	}

	return nil
}
