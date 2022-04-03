package broker

import (
	"strech-server/config"
	"strech-server/logger"
	"time"

	"github.com/nats-io/nats.go"
)

var configuration = config.GetConfig()

func handleDisconnectEvent(con *nats.Conn, err error) {
	logger.Error("Broker has disconnected: " + err.Error())
}

func handleAsyncErrors(con *nats.Conn, sub *nats.Subscription, err error) {
	logger.Error("Broker has experiences an error: " + err.Error())
}

func initializeBrokerConnection() *nats.Conn {
	options := nats.Options{
		Url:                  configuration.BROKER_URL,
		User:                 configuration.BROKER_INTERNAL_USER,
		Password:             configuration.BROKER_INTERNAL_PASSWORD,
		RetryOnFailedConnect: true,
		AllowReconnect:       true,
		MaxReconnect:         10,
		ReconnectWait:        5 * time.Second,
		Timeout:              1 * time.Second,
		PingInterval:         5 * time.Second,
		DisconnectedErrCB:    handleDisconnectEvent,
		AsyncErrorCB:         handleAsyncErrors,
	}
	nc, err := options.Connect()
	if err != nil {
		logger.Error("Failed to create connection with the broker: " + err.Error())
		panic("Failed to create connection with the broker: " + err.Error())
	}

	// js, err := nc.JetStream()
	// if err != nil {
	// 	return err
	// }

	return nc
}

func AddUser() error {
	return nil
}

func RemoveUser(username string) error {
	return nil
}

func CreateStream() error {
	return nil
}

func CreateProducer() error {
	return nil
}

func CreateConsumer() error {
	return nil
}

func RemoveStream() error {
	return nil
}

func RemoveProducer() error {
	return nil
}

func RemoveConsumer() error {
	return nil
}

func Close() {
	// broker.Close()
}

// var broker = initializeBrokerConnection()
