// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
