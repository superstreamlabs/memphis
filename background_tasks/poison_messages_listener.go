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
	"memphis-broker/logger"

	"github.com/nats-io/nats.go"
)

func poisonMessageHandler(msg *nats.Msg) {
	var message map[string]interface{}
	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		logger.Error("Error while getting notified about a poison message: " + err.Error())
	}

	// stationName := message["stream"].(string)
	// consumerName := message["consumer"].(string)
	// messageSeq := message["stream_seq"].(float64)
	// deliveriesCount := message["deliveries"].(float64)

	// msg := broker.GetMessage()
}

func ListenForPoisonMessages() {
	broker.QueueSubscribe("$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.>", "$memphis_poison_messages_listeners_group", poisonMessageHandler)
}
