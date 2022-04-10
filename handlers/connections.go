package handlers

import (
	"context"
	"memphis-control-plane/logger"
	"memphis-control-plane/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConnectionsHandler struct{}

func (umh ConnectionsHandler) CreateConnection(username string) (primitive.ObjectID, error) {
	connectionId := primitive.NewObjectID()
	newConnection := models.Connection{
		ID:            connectionId,
		CreatedByUser: username,
		CreationDate:  time.Now(),
	}

	_, err := connectionsCollection.InsertOne(context.TODO(), newConnection)
	if err != nil {
		logger.Error("CreateConnection error: " + err.Error())
		return connectionId, err
	}
	return connectionId, nil
}

func (umh ConnectionsHandler) GetConnection(connectionId primitive.ObjectID) (models.Connection, error) {
	var connection models.Connection
	err := connectionsCollection.FindOne(context.TODO(), bson.M{"_id": connectionId}).Decode(&connection)
	if err != nil {
		logger.Error("GetConnection error: " + err.Error())
		return connection, err
	}

	return connection, nil
}

func (umh ConnectionsHandler) GetAllConnections() ([]models.Connection, error) {
	var connections []models.Connection

	cursor, err := connectionsCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		logger.Error("GetAllConnections error: " + err.Error())
		return connections, err
	}

	if err = cursor.All(context.TODO(), &connections); err != nil {
		logger.Error("GetAllConnections error: " + err.Error())
		return connections, err
	}

	return connections, nil
}

func (umh ConnectionsHandler) RemoveConnection(connectionId primitive.ObjectID) error {
	_, err := connectionsCollection.DeleteOne(context.TODO(), bson.M{"_id": connectionId})
	if err != nil {
		logger.Error("RemoveConnection error: " + err.Error())
		return err
	}

	return nil
}
