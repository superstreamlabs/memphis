package notifications

import "go.mongodb.org/mongo-driver/mongo"

var NotificationIntegrationsMap map[string]interface{}
var IntegrationsCollection *mongo.Collection
