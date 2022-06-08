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

package background_tasks

import (
	"encoding/json"
	"memphis-broker/broker"
	"memphis-broker/handlers"
	"memphis-broker/logger"
	"memphis-broker/models"
	"memphis-broker/socketio"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

var sysLogsHandler = handlers.SysLogsHandler{}

func ConsumeSysLogs(wg *sync.WaitGroup) {
	defer wg.Done()

	sub, err := broker.CreatePullSubscriber("$memphis_sys_logs", "$memphis_sys_logs_consumers")
	if err != nil {
		panic("Failed creating sys log subscriber: " + err.Error())
	}

	for range time.Tick(time.Second * 30) {
		msgs, err := sub.Fetch(1000, nats.MaxWait(10*time.Second))

		if err != nil && !strings.Contains(err.Error(), "timeout") { // when subscriber done waiting and got no messages, we ignore timeout error
			logger.Error("Error fetching sys logs: " + err.Error())
		}

		if len(msgs) > 0 {
			logsForDB := make([]interface{}, len(msgs))
			logsForSocket := make([]models.SysLog, len(msgs))
			var singleLog models.SysLog

			for index, msg := range msgs {
				err := json.Unmarshal(msg.Data, &singleLog)
				if err != nil {
					logger.Error("Error converting sys logs: " + err.Error())
				} else {
					logsForDB[index] = singleLog
					logsForSocket[index] = singleLog
				}
			}

			err = sysLogsHandler.InsertLogs(logsForDB)
			if err != nil {
				logger.Error("Error inserting sys logs to DB: " + err.Error())
			}

			socketio.SendSysLogs(logsForSocket)
			for _, msg := range msgs {
				msg.Ack()
			}
		}
	}

}
