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
	"memphis-control-plane/models"
	"time"

	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var logger = log.Default()


func Info(logMessage string) {
	logger.Print("[INFO] " + logMessage)
	log := models.Log{
		ID: primitive.NewObjectID(),
		Log: logMessage,
		Type: "info",
		CreationDate: time.Now(),
		Component: "control-plane",
	}
	err := broker.PublishMessageToSubject("$memphis_sys_logs", log.ToBytes())
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func Warn(logMessage string) {
	logger.Print("[WARNING] " + logMessage)
	log := models.Log{
		ID: primitive.NewObjectID(),
		Log: logMessage,
		Type: "warn",
		CreationDate: time.Now(),
		Component: "control-plane",
	}
	err := broker.PublishMessageToSubject("$memphis_sys_logs", log.ToBytes())
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func Error(logMessage string) {
	logger.Print("[ERROR] " + logMessage)
	log := models.Log{
		ID: primitive.NewObjectID(),
		Log: logMessage,
		Type: "error",
		CreationDate: time.Now(),
		Component: "control-plane",
	}
	err := broker.PublishMessageToSubject("$memphis_sys_logs", log.ToBytes())
	if err != nil {
		logger.Print("[ERROR] Error saving logs: " + logMessage)
	}
}

func InitializeLogger() error{
	err := broker.CreateInternalStream("sys_logs")
	if err != nil {
		return err
	}
	return nil
}
