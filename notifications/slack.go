package notifications

import (
	"context"

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
	CacheSlackDetails(slackIntegration.Keys, slackIntegration.Properties)
	return nil
}

func clearSlackCache() {
	delete(NotificationIntegrationsMap, "slack")
}

func CacheSlackDetails(keys map[string]string, properties map[string]bool) {
	var authToken, channelID string
	var poisonMessageAlert, schemaValidationFailAlert, disconnectionEventsAlert bool
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
		disconnectionEventsAlert = false
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
	disconnectionEventsAlert, ok = properties["disconnection_events_alert"]
	if !ok {
		disconnectionEventsAlert = false
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
	slackIntegration.Properties["disconnection_events_alert"] = disconnectionEventsAlert
	slackIntegration.Name = "slack"
	NotificationIntegrationsMap["slack"] = slackIntegration
}

func SendMessageToSlackChannel(integration models.SlackIntegration, title string, message string) error {
	attachment := slack.Attachment{
		AuthorName: "Memphis",
		Title:      title,
		Text:       message,
		Color:      "#6557FF",
	}

	_, _, err := integration.Client.PostMessage(
		integration.Keys["channel_id"],
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return err
	}
	return nil
}
