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
	"log"
	"memphis-control-plane/config"
)

var logger = log.Default()
var configuration = config.GetConfig()

func Info(logMessage string) {
	logger.Print("[INFO] " + logMessage)
	// TODO send via socket
	// TODO  store in DB
}

func Warn(logMessage string) {
	logger.Print("[WARNING] " + logMessage)
	// TODO send via socket
	// TODO store in DB
}

func Error(logMessage string) {
	logger.Print("[ERROR] " + logMessage)
	// TODO send via socket
	// TODO store in DB
}
