package utils

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Filter struct
type Filter struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
}

// HeadersHandler func
func HeadersHandler(w http.ResponseWriter) {

}

// CreateUniqueIndex func
func CreateUniqueIndex(row string, collection *mongo.Collection) (string, error) {
	result, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{Keys: bsonx.Doc{{row, bsonx.Int32(1)}}, Options: options.Index().SetUnique(true)})
	return result, err
}
