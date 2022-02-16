package models

// "go.mongodb.org/mongo-driver/bson/primitive"

type Xxx struct {
	ID     string  `json:"id" bson:"_id"`
	Title  string  `json:"title" bson:"title"`
	Artist string  `json:"artist" bson:"artist"`
	Price  float64 `json:"price" bson:"price"`
}
