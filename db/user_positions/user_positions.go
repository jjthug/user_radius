package user_positions

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserPositionsRepo struct {
	MongoCollection *mongo.Collection
}

func (r *UserPositionsRepo) InsertUserLocationUpdate(user_loc UserLocation) (any, error) {
	result, err := r.MongoCollection.InsertOne(context.Background(), user_loc)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetNearbyLocations retrieves all records within a 1 km radius of the given latitude and longitude
func (r *UserPositionsRepo) GetNearbyUsers(ctx context.Context, latitude, longitude float64, distance int64) ([]UserLocation, error) {
	var results []UserLocation

	// Set radius to 1 km in meters (1000 meters)
	filter := bson.M{
		"location": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{longitude, latitude},
				},
				"$maxDistance": distance, // distance in meters
			},
		},
	}

	cursor, err := r.MongoCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby locations: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var record UserLocation
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("failed to decode location record: %w", err)
		}
		results = append(results, record)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	return results, nil
}
