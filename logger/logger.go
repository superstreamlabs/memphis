package logger

import (
	"log"
	"memphis-control-plane/config"
)

var logger = log.Default()
var configuration = config.GetConfig()

func Info(logMessage string) {
	logger.Print("[INFO] " + logMessage)
}

func Warn(logMessage string) {
	logger.Print("[WARNING] " + logMessage)
}

func Error(logMessage string) {
	logger.Print("[ERROR] " + logMessage)
}
