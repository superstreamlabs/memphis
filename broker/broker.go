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
// limitations under the License.

package broker

import (
	"memphis-broker/conf"
	"memphis-broker/models"
	"memphis-broker/server"
	"strings"
	"time"

	"errors"
	"log"

	// "strings"
	// "time"

	// "github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	// "github.com/nats-io/nkeys"
)

var configuration = conf.GetConfig()
var connectionChannel = make(chan bool)
var logger = log.Default()

// func getErrorWithoutNats(err error) error {
// 	// message := strings.ToLower(err.Error())
// 	// message = strings.Replace(message, "nats", "memphis-broker", -1)
// 	// return errors.New(message)
// 	return errors.New("not implemented")
// }

// func handleDisconnectEvent(con *nats.Conn, err error) {
// 	// if err != nil {
// 	// 	logger.Print("[Error] Broker has disconnected: " + err.Error())
// 	// }
// }

// func handleAsyncErrors(con *nats.Conn, sub *nats.Subscription, err error) {
// 	// logger.Print("[Error] Broker has experienced an error: " + err.Error())
// }

func handleReconnect(con *nats.Conn) {
	// if connected {
	// 	logger.Print("[INFO] Reconnected to the broker")
	// }
	// connectionChannel <- true
}

func handleClosed(con *nats.Conn) {
	// if !connected {
	// 	logger.Print("[INFO] All reconnect attempts with the broker were failed")
	// 	connectionChannel <- false
	// }
}

func sigHandler(nonce []byte, seed string) ([]byte, error) {
	// kp, err := nkeys.FromSeed([]byte(seed))
	// if err != nil {
	// 	return nil, err
	// }

	// defer kp.Wipe()

	// sig, _ := kp.Sign(nonce)
	// return sig, nil
	return nil, errors.New("not implemented")
}

func userCredentials(userJWT string, userKeySeed string) nats.Option {
	// userCB := func() (string, error) {
	// 	return userJWT, nil
	// }
	// sigCB := func(nonce []byte) ([]byte, error) {
	// 	return sigHandler(nonce, userKeySeed)
	// }
	// return nats.UserJWT(userCB, sigCB)
	return nil
}

func initializeBrokerConnection() (*nats.Conn, nats.JetStreamContext) {
	// nc, err := nats.Connect(
	// 	configuration.BROKER_URL,
	// 	// nats.UserCredentials("admin3.creds"),
	// 	// userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
	// 	nats.Token(configuration.CONNECTION_TOKEN),
	// 	nats.RetryOnFailedConnect(true),
	// 	nats.MaxReconnects(10),
	// 	nats.ReconnectWait(5*time.Second),
	// 	nats.Timeout(10*time.Second),
	// 	nats.PingInterval(5*time.Second),
	// 	nats.DisconnectErrHandler(handleDisconnectEvent),
	// 	nats.ErrorHandler(handleAsyncErrors),
	// 	nats.ReconnectHandler(handleReconnect),
	// 	nats.ClosedHandler(handleClosed),
	// )

	// if !nc.IsConnected() {
	// 	isConnected := <-connectionChannel
	// 	if !isConnected {
	// 		logger.Print("[Error] Failed to create connection with the broker")
	// 		panic("Failed to create connection with the broker")
	// 	}
	// }

	// if err != nil {
	// 	logger.Print("[Error] Failed to create connection with the broker: " + err.Error())
	// 	panic("Failed to create connection with the broker: " + err.Error())
	// }

	// js, err := nc.JetStream()
	// if err != nil {
	// 	logger.Print("[Error] Failed to create connection with the broker: " + err.Error())
	// 	panic("Failed to create connection with the broker: " + err.Error())
	// }

	// connected = true
	// // logger.Print("[INFO] Established connection with the broker")
	// return nc, js

	return nil, nil
}

func AddUser(username string) (string, error) {
	return configuration.CONNECTION_TOKEN, nil
}

func RemoveUser(username string) error {
	return nil
}

func CreateStream(s *server.Server, station models.Station) error {
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

	var storage server.StorageType
	if station.StorageType == "memory" {
		storage = server.MemoryStorage
	} else {
		storage = server.FileStorage
	}

	var dedupWindow time.Duration
	if station.DedupEnabled && station.DedupWindowInMs >= 100 {
		dedupWindow = time.Duration(station.DedupWindowInMs) * time.Millisecond
	} else {
		dedupWindow = time.Duration(100) * time.Millisecond // can not be 0
	}

	return s.MemphisAddStream(&server.StreamConfig{
		Name:         station.Name,
		Subjects:     []string{station.Name + ".>"},
		Retention:    server.LimitsPolicy,
		MaxConsumers: -1,
		MaxMsgs:      int64(maxMsgs),
		MaxBytes:     int64(maxBytes),
		Discard:      server.DiscardOld,
		MaxAge:       maxAge,
		MaxMsgsPer:   -1,
		MaxMsgSize:   int32(configuration.MAX_MESSAGE_SIZE_MB) * 1024 * 1024,
		Storage:      storage,
		Replicas:     station.Replicas,
		NoAck:        false,
		Duplicates:   dedupWindow,
	})
}

func CreateProducer() error {
	// nothing to create
	return nil
}

func CreateConsumer(s *server.Server, consumer models.Consumer, station models.Station) error {
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

	err := s.MemphisAddConsumer(station.Name, &server.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: server.DeliverAll,
		AckPolicy:     server.AckExplicit,
		AckWait:       time.Duration(maxAckTimeMs) * time.Millisecond,
		MaxDeliver:    MaxMsgDeliveries,
		FilterSubject: station.Name + ".final",
		ReplayPolicy:  server.ReplayInstant,
		MaxAckPending: -1,
		HeadersOnly:   false,
		// RateLimit: ,// Bits per sec
		// Heartbeat: // time.Duration,
	})
	return err
}

func GetCgInfo(s *server.Server, stationName, cgName string) (*server.ConsumerInfo, error) {
	info, err := s.MemphisGetConsumerInfo(stationName, cgName)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func RemoveStream(s *server.Server, streamName string) error {
	return s.MemphisRemoveStream(streamName)
}

func GetTotalMessagesInStation(s *server.Server, station models.Station) (int, error) {
	streamInfo, err := s.MemphisStreamInfo(station.Name)
	if err != nil {
		return 0, err
	}

	return int(streamInfo.State.Msgs), nil
}

func GetTotalMessagesAcrossAllStations(s *server.Server) (int, error) {
	messagesCounter := 0
	for _, streamInfo := range s.MemphisAllStreamsInfo() {
		if !strings.HasPrefix(streamInfo.Config.Name, "$memphis") { // skip internal streams
			messagesCounter = messagesCounter + int(streamInfo.State.Msgs)
		}
	}

	return messagesCounter, nil
}

func GetAvgMsgSizeInStation(s *server.Server, station models.Station) (int64, error) {
	streamInfo, err := s.MemphisStreamInfo(station.Name)
	if err != nil || streamInfo.State.Bytes == 0 {
		return 0, err
	}

	return int64(streamInfo.State.Bytes / streamInfo.State.Msgs), nil
}

func GetHeaderSizeInBytes(headers nats.Header) int {
	// bytes := 0
	// for i, s := range headers {
	// 	bytes += len(s[0]) + len(i)
	// }
	// return bytes
	return 0
}

func GetMessages(station models.Station, messagesToFetch int) ([]models.MessageDetails, error) {
	// streamInfo, err := js.StreamInfo(station.Name)
	// if err != nil {
	// 	return []models.MessageDetails{}, getErrorWithoutNats(err)
	// }
	// totalMessages := streamInfo.State.Msgs

	// var startSequence uint64 = 1
	// if totalMessages > uint64(messagesToFetch) {
	// 	startSequence = totalMessages - uint64(messagesToFetch) + 1
	// }

	// uid, _ := uuid.NewV4()
	// durableName := "$memphis_fetch_messages_consumer" + uid.String()
	// sub, err := js.PullSubscribe(station.Name+".final", durableName, nats.StartSequence(startSequence))
	// msgs, _ := sub.Fetch(messagesToFetch, nats.MaxWait(3*time.Second))
	// var messages []models.MessageDetails
	// for _, msg := range msgs {
	// 	if msg.Header.Get("producedBy") == "$memphis_dlq" { // skip poison messages which have been resent
	// 		continue
	// 	}

	// 	metadata, _ := msg.Metadata()
	// 	data := (string(msg.Data))
	// 	if len(data) > 100 { // get the first chars for preview needs
	// 		data = data[0:100]
	// 	}
	// 	messages = append(messages, models.MessageDetails{
	// 		MessageSeq:   int(metadata.Sequence.Stream),
	// 		Data:         data,
	// 		ProducedBy:   msg.Header.Get("producedBy"),
	// 		ConnectionId: msg.Header.Get("connectionId"),
	// 		TimeSent:     metadata.Timestamp,
	// 		Size:         len(msg.Subject) + len(msg.Data) + GetHeaderSizeInBytes(msg.Header),
	// 	})
	// 	msg.Ack()
	// }

	// for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 { // sort from new to old
	// 	messages[i], messages[j] = messages[j], messages[i]
	// }

	// js.DeleteConsumer(station.Name, durableName)
	// return messages, nil
	return nil, errors.New("not implemented")
}

func GetMessage(stationName string, messageSeq uint64) (*nats.RawStreamMsg, error) {
	// msg, err := js.GetMsg(stationName, messageSeq)
	// if err != nil {
	// 	return nil, err
	// }

	// return msg, nil
	return nil, errors.New("not implemented")
}

func ResendPoisonMessage(subject string, data []byte) error {
	// natsMessage := &nats.Msg{
	// 	Header:  map[string][]string{"producedBy": {"$memphis_dlq"}},
	// 	Subject: subject,
	// 	Data:    data,
	// }

	// err := broker.PublishMsg(natsMessage)
	// if err != nil {
	// 	return err
	// }

	// return nil
	return errors.New("not implemented")
}

func RemoveProducer() error {
	// // nothing to remove
	// return nil
	return errors.New("not implemented")
}

func RemoveConsumer(streamName string, consumerName string) error {
	// err := js.DeleteConsumer(streamName, consumerName)
	// if err != nil {
	// 	return getErrorWithoutNats(err)
	// }

	// return nil
	return errors.New("not implemented")
}

func ValidateUserCreds(token string) error {
	// nc, err := nats.Connect(
	// 	configuration.BROKER_URL,
	// 	// nats.UserCredentials("admin3.creds"),
	// 	// userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
	// 	nats.Token(token),
	// )

	// if err != nil {
	// 	return getErrorWithoutNats(err)
	// }

	// _, err = nc.JetStream()
	// if err != nil {
	// 	return getErrorWithoutNats(err)
	// }

	// nc.Close()
	// return nil
	return errors.New("not implemented")
}

func CreateInternalStream(s *server.Server, name string) error {
	return s.MemphisAddStream(&server.StreamConfig{
		Name:         name,
		Subjects:     []string{name},
		Retention:    server.WorkQueuePolicy,
		MaxConsumers: -1,
		Storage:      server.FileStorage,
		Replicas:     1,
		NoAck:        false,
		Duplicates:   100 * time.Millisecond,
	})
}

func PublishMessageToSubject(subject string, msg []byte) error {
	// _, err := js.Publish(subject, msg)
	// if err != nil {
	// 	return getErrorWithoutNats(err)
	// }
	// return nil
	return errors.New("not implemented")
}

func CreatePullSubscriber(stream string, durable string) (*nats.Subscription, error) {
	// sub, err := js.PullSubscribe(stream, durable)
	// if err != nil {
	// 	return sub, getErrorWithoutNats(err)
	// }
	// return sub, nil
	return nil, errors.New("not implemented")
}

func QueueSubscribe(subject, queue_group_name string, cb func(msg *nats.Msg)) {
	// broker.QueueSubscribe(subject, queue_group_name, cb)
}

func IsConnectionAlive() bool {
	// return broker.IsConnected()
	return false
}

func Close() {
	// broker.Close()
}

var broker, js = initializeBrokerConnection()
