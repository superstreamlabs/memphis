package slack

import (
	"context"

	"memphis-broker/db"
	"memphis-broker/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/slack-go/slack"
)

var AuthToken string
var ChannelID string
var PoisonMessageAlert bool
var SchemaValidationFailAlert bool
var SlackClient *slack.Client
var integrationsCollection *mongo.Collection

func InitializeSlackConnection(c *mongo.Client) error {
	integrationsCollection = db.GetCollection("integrations", c)
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := integrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	UpdateSlackDetails(slackIntegration.Keys["auth_token"], slackIntegration.Keys["channel_id"], slackIntegration.Properties["poison_message_alert"], slackIntegration.Properties["schema_validation_fail_alert"])
	return nil
}

func UpdateSlackDetails(authToken string, channelID string, poisonMessageAlert bool, schemaValidationFailAlert bool) {
	if AuthToken != authToken {
		AuthToken = authToken
		if authToken != "" {
			SlackClient = slack.New(authToken)
		}
	}
	ChannelID = channelID
	PoisonMessageAlert = poisonMessageAlert
	SchemaValidationFailAlert = schemaValidationFailAlert
}

func SendMessageToSlackChannel(message string) error {
	attachment := slack.Attachment{
		Pretext: "Memphis Notification",
		Text:    message,
		Color:   "#6557FF",
	}
	_, _, err := SlackClient.PostMessage(
		ChannelID,
		slack.MsgOptionAttachments(attachment),
	)

	if err != nil {
		return err
	}
	return nil
}
