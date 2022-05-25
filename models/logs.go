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

package models

import (
	"bytes"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Log struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Log          string             `json:"log" bson:"log"`
	Type         string             `json:"type" bson:"type"`
	CreationDate time.Time          `json:"creation_date" bson:"creation_date"`
}

func (log Log) ToBytes() []byte {
	bytesLog := new(bytes.Buffer)
	json.NewEncoder(bytesLog).Encode(log)
	return bytesLog.Bytes()
}
