package broker

import (
	"errors"
	"memphis-control-plane/config"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

var configuration = config.GetConfig()

func getErrorWithoutNats(err error) error {
	message := strings.ToLower(err.Error())
	message = strings.Replace(message, "nats", "mmphis-broker", -1)
	return errors.New(message)
}

func handleDisconnectEvent(con *nats.Conn, err error) {
	logger.Error("Broker has disconnected: " + err.Error())
}

func handleAsyncErrors(con *nats.Conn, sub *nats.Subscription, err error) {
	logger.Error("Broker has experiences an error: " + err.Error())
}

func sigHandler(nonce []byte, seed string) ([]byte, error) {
	kp, err := nkeys.FromSeed([]byte(seed))
	if err != nil {
		return nil, err
	}

	defer kp.Wipe()

	sig, _ := kp.Sign(nonce)
	return sig, nil
}

func userCredentials(userJWT string, userKeySeed string) nats.Option {
	userCB := func() (string, error) {
		return userJWT, nil
	}
	sigCB := func(nonce []byte) ([]byte, error) {
		return sigHandler(nonce, userKeySeed)
	}
	return nats.UserJWT(userCB, sigCB)
}

func initializeBrokerConnection() (*nats.Conn, nats.JetStreamContext) {
	nc, err := nats.Connect(
		configuration.BROKER_URL,
		// nats.UserCredentials("admin3.creds"),
		// userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
		nats.Token(configuration.CONNECTION_TOKEN),
		nats.MaxReconnects(10),
		nats.ReconnectWait(5*time.Second),
		nats.Timeout(10*time.Second),
		nats.PingInterval(5*time.Second),
		nats.DisconnectErrHandler(handleDisconnectEvent),
		nats.ErrorHandler(handleAsyncErrors),
	)

	if err != nil {
		logger.Error("Failed to create connection with the broker: " + err.Error())
		panic("Failed to create connection with the broker: " + err.Error())
	}

	js, err := nc.JetStream()
	if err != nil {
		logger.Error("Failed to create connection with the broker: " + err.Error())
		panic("Failed to create connection with the broker: " + err.Error())
	}

	return nc, js
}

func AddUser(username string) (string, error) {
	return configuration.CONNECTION_TOKEN, nil
}

func RemoveUser(username string) error {
	return nil
}

func CreateStream(station models.Station) error {
	var maxMsgs int
	if station.RetentionType == "messages" && station.RetentionValue > 0 {
		maxMsgs = station.RetentionValue
	} else {
		maxMsgs = -1
	}

	var maxByes int
	if station.RetentionType == "bytes" && station.RetentionValue > 0 {
		maxByes = station.RetentionValue
	} else {
		maxByes = -1
	}

	var maxAge time.Duration
	if station.RetentionType == "message_age_sec" && station.RetentionValue > 0 {
		maxAge = time.Duration(station.RetentionValue) * time.Second
	} else {
		maxAge = time.Duration(-1)
	}

	var storage nats.StorageType
	if station.StorageType == "memory" {
		storage = nats.MemoryStorage
	} else {
		storage = nats.FileStorage
	}

	var dedupWindow time.Duration
	if station.DedupEnabled {
		dedupWindow = time.Duration(station.DedupWindowInMs*1000) * time.Nanosecond
	} else {
		dedupWindow = time.Duration(0)
	}

	_, err := js.AddStream(&nats.StreamConfig{
		Name:              station.Name,
		Subjects:          []string{station.Name + ".*"},
		Retention:         nats.LimitsPolicy,
		MaxConsumers:      -1,
		MaxMsgs:           int64(maxMsgs),
		MaxBytes:          int64(maxByes),
		Discard:           nats.DiscardOld,
		MaxAge:            maxAge,
		MaxMsgsPerSubject: -1,
		MaxMsgSize:        int32(configuration.MAX_MESSAGE_SIZE_MB) * 1024,
		Storage:           storage,
		Replicas:          station.Replicas,
		NoAck:             false,
		Duplicates:        dedupWindow,
	}, nats.MaxWait(15*time.Second))
	if err != nil {
		return getErrorWithoutNats(err)
	}

	return nil
}

func CreateProducer() error {
	// nothing to create
	return nil
}

func CreateConsumer(consumer models.Consumer, station models.Station) error {
	var consumerName string
	if consumer.ConsumersGroup != "" {
		consumerName = consumer.ConsumersGroup + "_group"
	} else {
		consumerName = consumer.Name
	}

	_, err := js.AddConsumer(station.Name, &nats.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       time.Duration(30*1000*1000) * time.Nanosecond, // 30 sec
		MaxDeliver:    10,
		FilterSubject: station.Name + ".final",
		ReplayPolicy:  nats.ReplayInstantPolicy,
		MaxAckPending: -1,
		HeadersOnly:   false,
		// RateLimit: ,// Bits per sec
		// Heartbeat: // time.Duration,
	})
	if err != nil {
		return getErrorWithoutNats(err)
	}

	return nil
}

func RemoveStream(streamName string) error {
	err := js.DeleteStream(streamName)
	if err != nil {
		return err
	}

	return nil
}

func RemoveProducer() error {
	// nothing to remove
	return nil
}

func RemoveConsumer(streamName string, consumerName string) error {
	err := js.DeleteConsumer(streamName, consumerName)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUserCreds(token string) error {
	nc, err := nats.Connect(
		configuration.BROKER_URL,
		// nats.UserCredentials("admin3.creds"),
		// userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
		nats.Token(token),
	)

	if err != nil {
		return getErrorWithoutNats(err)
	}

	_, err = nc.JetStream()
	if err != nil {
		return getErrorWithoutNats(err)
	}

	nc.Close()
	return nil
}

func Close() {
	broker.Close()
}

var broker, js = initializeBrokerConnection()
