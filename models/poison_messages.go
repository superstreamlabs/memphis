package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessagePayload struct {
	ConnectionId    primitive.ObjectID `json:"connection_id" bson:"connection_id"`
	ProducedBy      string             `json:"produced_by" bson:"produced_by"`
	ProducerAddress string             `json:"producer_address" bson:"producer_address"`
	TimeProduced    time.Time          `json:"time_produced" bson:"time_produced"`
	Size            int                `json:"size" bson:"size"`
	Data            string             `json:"data" bson:"data"`
}

type PoisonedCg struct {
	CgName              string     `json:"cg_name" bson:"cg_name"`
	PoisoningTime       time.Time  `json:"poisoning_time" bson:"poisoning_time"`
	DeliveriesCount     int        `json:"deliveries_count" bson:"deliveries_count"`
	UnprocessedMessages int        `json:"unprocessed_messages" bson:"unprocessed_messages"`
	MaxAckTimeMs        int64      `json:"max_ack_time_ms" bson:"max_ack_time_ms"`
	InProcessMessages   int        `json:"in_process_messages" bson:"in_process_messages"`
	TotalPoisonMessages int        `json:"total_poison_messages" bson:"total_poison_messages"`
	MaxMsgDeliveries    int        `json:"max_msg_deliveries" bson:"max_msg_deliveries"`
	CgMembers           []CgMember `json:"cg_members" bson:"cg_members"`
}

type PoisonMessage struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	StationName string             `json:"station_name" bson:"station_name"`
	MessageSeq  int                `json:"message_seq" bson:"message_seq"`
	Cgs         []PoisonedCg       `json:"cgs" bson:"cgs"`
	Message     MessagePayload     `json:"message" bson:"message"`
}

type LightweightPoisonMessage struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Data string             `json:"data" bson:"data"`
}
