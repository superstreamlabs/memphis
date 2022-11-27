package notifications

import (
	"errors"
	"memphis-broker/models"

	"go.mongodb.org/mongo-driver/mongo"
)

var NotificationIntegrationsMap map[string]interface{}
var NotificationFunctionsMap map[string]interface{}
var IntegrationsCollection *mongo.Collection

const PoisonMAlert = "poison_message_alert"
const SchemaVAlert = "schema_validation_fail_alert"
const DisconEAlert = "disconnection_events_alert"

func SendNotification(title string, message string, msgType string) error {
	for k, f := range NotificationFunctionsMap {
		switch k {
		case "slack":
			slackIntegration, ok := NotificationIntegrationsMap["slack"].(models.SlackIntegration)
			if ok {
				if slackIntegration.Properties[msgType] {
					err := f.(func(models.SlackIntegration, string, string) error)(slackIntegration, title, message)
					if err != nil {
						return err
					}
				}
			}
		default:
			return errors.New("Failed sending notification: unsupported integration")
		}
	}
	return nil

}
