package user_positions

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserLocation struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"user_id"`
	Location primitive.M        `bson:"location"` // Store latitude and longitude in GeoJSON format
}
