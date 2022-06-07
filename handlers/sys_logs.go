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

package handlers

import (
	"context"
	"memphis-broker/db"
	"memphis-broker/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sysLogsCollection *mongo.Collection = db.GetCollection("system_logs")

type SysLogsHandler struct{}

func (lh SysLogsHandler) GetSysLogs(hours int) ([]models.SysLog, error) {
	var logs []models.SysLog
	filter := bson.M{"creation_date": bson.M{"$gte": (time.Now().Add(-(time.Hour * time.Duration(hours))))}}
	opts := options.Find().SetSort(bson.D{{"creation_date", -1}})
	cursor, err := sysLogsCollection.Find(context.TODO(), filter, opts)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}

	return logs, nil
}

func (lh SysLogsHandler) InsertLogs(logs []interface{}) error {
	_, err := sysLogsCollection.InsertMany(context.TODO(), logs)
	if err != nil {
		return err
	}

	return nil
}
