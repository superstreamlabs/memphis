package logger

import (
	"log"
	"strech-server/config"
)

var logger = log.Default()
var configuration = config.GetConfig()

func Info(logMessage string) {
	destination := configuration.LOGGER
	switch destination {
	case "STDOUT":
		logger.Print(logMessage)
	}
}

func Error(logMessage string) {
	destination := configuration.LOGGER
	switch destination {
	case "STDOUT":
		logger.Print(logMessage)
	}
}