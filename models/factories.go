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
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Factory struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
}

type ExtendedFactory struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	CreatedByUser string             `json:"created_by_user" bson:"created_by_user"`
	CreationDate  time.Time          `json:"creation_date" bson:"creation_date"`
	UserAvatarId  int                `json:"user_avatar_id" bson:"user_avatar_id"`
}

type CreateFactorySchema struct {
	Name        string `json:"name" binding:"required,min=1,max=25"`
	Description string `json:"description"`
}

type GetFactorySchema struct {
	FactoryName string `form:"factory_name" json:"factory_name"  binding:"required"`
}

type RemoveFactorySchema struct {
	FactoryName string `json:"factory_name"  binding:"required"`
}

type EditFactorySchema struct {
	FactoryName    string `json:"factory_name"  binding:"required"`
	NewName        string `json:"factory_new_name"`
	NewDescription string `json:"factory_new_description"`
}
