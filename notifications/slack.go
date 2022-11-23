package notifications

import (
	"context"
	"errors"

	"memphis-broker/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/slack-go/slack"
)

func InitializeSlackConnection(c *mongo.Client) error {
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	CacheSlackDetails(slackIntegration.Keys, slackIntegration.Properties, slackIntegration.UIUrl)
	return nil
}

func clearSlackCache() {
	delete(NotificationIntegrationsMap, "slack")
}

func CacheSlackDetails(keys map[string]string, properties map[string]bool, url string) {
	var authToken, channelID string
	var poisonMessageAlert, schemaValidationFailAlert bool
	var slackIntegration models.SlackIntegration

	slackIntegration, ok := NotificationIntegrationsMap["slack"].(models.SlackIntegration)
	if !ok {
		slackIntegration = models.SlackIntegration{}
		slackIntegration.Keys = make(map[string]string)
		slackIntegration.Properties = make(map[string]bool)
	}
	if keys == nil {
		clearSlackCache()
		return
	}
	if properties == nil {
		poisonMessageAlert = false
		schemaValidationFailAlert = false
	}
	authToken, ok = keys["auth_token"]
	if !ok {
		clearSlackCache()
		return
	}
	channelID, ok = keys["channel_id"]
	if !ok {
		clearSlackCache()
		return
	}
	poisonMessageAlert, ok = properties["poison_message_alert"]
	if !ok {
		poisonMessageAlert = false
	}
	schemaValidationFailAlert, ok = properties["schema_validation_fail_alert"]
	if !ok {
		schemaValidationFailAlert = false
	}
	if slackIntegration.Keys["auth_token"] != authToken {
		slackIntegration.Keys["auth_token"] = authToken
		if authToken != "" {
			slackIntegration.Client = slack.New(authToken)
		}
	}

	slackIntegration.Keys["channel_id"] = channelID
	slackIntegration.Properties["poison_message_alert"] = poisonMessageAlert
	slackIntegration.Properties["schema_validation_fail_alert"] = schemaValidationFailAlert
	slackIntegration.Name = "slack"
	slackIntegration.UIUrl = url
	NotificationIntegrationsMap["slack"] = slackIntegration
}

func SendMessageToSlackChannel(title string, message string) error {
	slackIntegration, ok := NotificationIntegrationsMap["slack"].(models.SlackIntegration)
	if ok {
		attachment := slack.Attachment{
			Pretext: "Memphis",
			Title:   title,
			Text:    message,
			Color:   "#6557FF",
		}

		_, _, err := slackIntegration.Client.PostMessage(
			slackIntegration.Keys["channel_id"],
			slack.MsgOptionAttachments(attachment),
		)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Invalid slack credentials")
}
