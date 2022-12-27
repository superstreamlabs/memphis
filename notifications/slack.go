package notifications

import (
	"context"

	"memphis-broker/models"

	"github.com/slack-go/slack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func IsSlackEnabled() (bool, error) {
	filter := bson.M{"name": "slack"}
	var slackIntegration models.Integration
	err := IntegrationsCollection.FindOne(context.TODO(),
		filter).Decode(&slackIntegration)
	if err == nil {
		return true, nil
	}

	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return false, err
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
	poisonMessageAlert, ok = properties[PoisonMAlert]
	if !ok {
		poisonMessageAlert = false
	}
	schemaValidationFailAlert, ok = properties[SchemaVAlert]
	if !ok {
		schemaValidationFailAlert = false
	}
	disconnectionEventsAlert, ok = properties[DisconEAlert]
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
	slackIntegration.Properties[PoisonMAlert] = poisonMessageAlert
	slackIntegration.Properties[SchemaVAlert] = schemaValidationFailAlert
	slackIntegration.Properties[DisconEAlert] = disconnectionEventsAlert
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
