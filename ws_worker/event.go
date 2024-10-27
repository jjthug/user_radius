package ws_worker

import (
	"context"
	"encoding/json"
	"fmt"
	"homo_hunter_backend/db/user_positions"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, m *Manager) error

const (
	// EventSendMessage is the event name for new chat messages sent
	EventLocationUpdate = "location_update"
)

// SendMessageEvent is the payload sent in the
// send_message event
type LocationUpdateEvent struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	From      string `json:"from"`
}

// LocationUpdateHandler will send out a message to all other participants
func LocationUpdateHandler(event Event, m *Manager) error {
	// Marshal Payload into wanted format
	var locationUpdateEvent LocationUpdateEvent
	if err := json.Unmarshal(event.Payload, &locationUpdateEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	//push this to mongodb locationupdate collection
	userLocation := user_positions.UserLocation{
		UserID: locationUpdateEvent.From,
		Location: primitive.M{
			"type":        "Point",
			"coordinates": []interface{}{locationUpdateEvent.Longitude, locationUpdateEvent.Latitude},
		},
	}
	if _, err := m.repo.InsertUserLocationUpdate(userLocation); err != nil {
		return fmt.Errorf("failed to insert user location: %v", err)
	}

	// push this event to redis pub sub
	ctx := context.Background()
	return m.DistributeTaskLocationUpdate(ctx, &locationUpdateEvent)
}
