package logger

import (
	"log"
	"memphis-control-plane/config"
)

var logger = log.Default()
var configuration = config.GetConfig()

func Info(logMessage string) {
	logger.Print(logMessage)
}

func Error(logMessage string) {
	logger.Print(logMessage)
}
