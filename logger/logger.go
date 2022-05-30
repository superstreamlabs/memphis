// Copyright 2021-2022 The Memphis Authors
// Licensed under the GNU General Public License v3.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"memphis-control-plane/broker"
	"memphis-control-plane/config"

	"log"
)

var configuration = config.GetConfig()
var connectionChannel = make(chan bool)
var connected = false
var logger = log.Default()



func Info(logMessage string) {
	logger.Print("[INFO] " + logMessage)
	// TODO send via socket
	// TODO  store in DB
	err := broker.PublishLogToStream("$memphis_sys_logs", logMessage, "info", "control-plane")
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func Warn(logMessage string) {
	logger.Print("[WARNING] " + logMessage)
	// TODO send via socket
	// TODO store in DB
	err := broker.PublishLogToStream("$memphis_sys_logs", logMessage, "warn", "control-plane")
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func Error(logMessage string) {
	logger.Print("[ERROR] " + logMessage)
	// TODO send via socket
	// TODO store in DB
	err := broker.PublishLogToStream("$memphis_sys_logs", logMessage, "error", "control-plane")
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func InitializeLogger() error{
	err := broker.CreateInternalStream("sys_logs", []string{"$memphis_sys_logs"})
	if err != nil {
		return err
	}
	return nil
}
