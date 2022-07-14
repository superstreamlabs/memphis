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

type SandboxUser struct {
	ID              primitive.ObjectID `json:"id" bson:"_id"`
	Username        string             `json:"username" bson:"username"`
	Email           string             `json:"email" bson:"email"`
	FirstName       string             `json:"first_name" bson:"first_name"`
	LastName        string             `json:"last_name" bson:"last_name"`
	Password        string             `json:"password" bson:"password"`
	HubUsername     string             `json:"hub_username" bson:"hub_username"`
	HubPassword     string             `json:"hub_password" bson:"hub_password"`
	UserType        string             `json:"user_type" bson:"user_type"`
	AlreadyLoggedIn bool               `json:"already_logged_in" bson:"already_logged_in"`
	CreationDate    time.Time          `json:"creation_date" bson:"creation_date"`
	AvatarId        int                `json:"avatar_id" bson:"avatar_id"`
	ProfilePic      string             `json:"profile_pic" bson:"profile_pic"`
}

type SandboxLoginSchema struct {
	LoginType string `json:"login_type" binding:"required"`
	Token     string `json:"token" binding:"required"`
}
