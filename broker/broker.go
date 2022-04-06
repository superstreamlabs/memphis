package broker

import (
	"strech-server/config"
	"strech-server/logger"
	"strech-server/models"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

var configuration = config.GetConfig()

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
		userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
		nats.RetryOnFailedConnect(true),
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

func AddUser(username string) error {
	return nil
}

func RemoveUser(username string) error {
	return nil
}

func CreateStream(station models.Station) error {

	return nil
}

func CreateProducer() error {
	return nil
}

func CreateConsumer() error {
	return nil
}

func RemoveStream(stationName string) error {

	return nil
}

func RemoveProducer() error {
	return nil
}

func RemoveConsumer() error {
	return nil
}

func Close() {
	broker.Close()
}

var broker, js = initializeBrokerConnection()
