package broker

import (
	// "github.com/nats-io/nats.go"
	"strech-server/config"
	// "strech-server/logger"
)

var configuration = config.GetConfig()

func CreateStream() error {
	// nc, err := nats.Connect(configuration.BROKER_URL)
	// if err != nil {
	// 	return err
	// }

	// js, err := nc.JetStream()
	// if err != nil {
	// 	return err
	// }
	return nil

}
