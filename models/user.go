package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Username     string             `json:"username" binding:"required,min=1,max=25" bson:"username"`
	Password     string             `json:"password" binding:"required,min=6" bson:"password"`
	HubUsername  string             `json:"hub_username" bson:"hub_username"`
	HubPassword  string             `json:"hub_password" bson:"hub_password"`
	UserType     string             `json:"user_type" binding:"required" bson:"user_type"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
}

func (u *User) Validate(c *gin.Context) (validator.ValidationErrors, bool) {
	if err := c.ShouldBindJSON(u); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return verr, true
		}
		c.AbortWithStatusJSON(400, gin.H{"message": "Body params have to be in JSON format"})
		return nil, true
	}

	return nil, false
}