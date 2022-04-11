package broker

import (
	"encoding/base64"
	"memphis-control-plane/config"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
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
	token, _ := base64.StdEncoding.DecodeString(configuration.CONNECTION_TOKEN)

	nc, err := nats.Connect(
		configuration.BROKER_URL,
		// nats.UserCredentials("admin3.creds"),
		// userCredentials(configuration.BROKER_ADMIN_JWT, configuration.BROKER_ADMIN_NKEY),
		nats.Token(string(token)),
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
	token, _ := base64.StdEncoding.DecodeString(configuration.CONNECTION_TOKEN)
	return string(token), nil
}

func RemoveUser(username string) error {
	return nil
}

func CreateStream(station models.Station) error {
	// x, err := js.AddStream(&nats.StreamConfig{
	// 	Name:     "ORDERS",
	// 	Subjects: []string{"ORDERS.*"},
	// }, nats.MaxWait(15*time.Second))
	// if err != nil || x == nil {
	// 	logger.Error("Failed to create connection with the broker: " + err.Error())
	// }
	return nil
}

func CreateProducer() error {
	return nil
}

func CreateConsumer() error {
	// js.AddConsumer("ORDERS", &nats.ConsumerConfig{
	// 	Durable: "MONITOR",
	// })

	return nil
}

func RemoveStream(stationName string) error {
	// err := js.DeleteStream("ORDERS")
	// if err != nil {
	// 	logger.Error("Failed to create connection with the broker: " + err.Error())
	// }
	return nil
}

func RemoveProducer() error {
	return nil
}

func RemoveConsumer() error {
	// js.DeleteConsumer("ORDERS", "MONITOR")
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
		return err
	}

	_, err = nc.JetStream()
	if err != nil {
		return err
	}

	nc.Close()
	return nil
}

func Close() {
	broker.Close()
}

var broker, js = initializeBrokerConnection()
