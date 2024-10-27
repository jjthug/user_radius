package ws_worker

import (
	"context"
	"errors"
	"homo_hunter_backend/db/user_positions"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hibiken/asynq"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

var (
	/**
	websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection
	*/
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
	clients ClientList
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers        map[string]EventHandler
	taskDistributor RedisTaskDistributor
	taskProcessor   RedisTaskProcessor
	repo            user_positions.UserPositionsRepo
}

// NewManager is used to initalize all the values inside the manager
func NewManager(ctx context.Context, redisOpt asynq.RedisClientOpt) *Manager {
	m := &Manager{
		clients:         make(ClientList),
		handlers:        make(map[string]EventHandler),
		taskDistributor: *NewRedisTaskDistributor(redisOpt),
		taskProcessor:   *NewRedisTaskProcessor(redisOpt),
	}
	m.setupEventHandlers()
	return m
}

// setupEventHandlers configures and adds all handlers
func (m *Manager) setupEventHandlers() {
	m.handlers[EventLocationUpdate] = LocationUpdateHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *Manager) routeEvent(event Event) error {
	// Check if Handler is present in Map
	if handler, ok := m.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, m); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

// serveWS is a HTTP Handler that the has the Manager that allows connections
func (m *Manager) ServeWS(c *gin.Context) {
	log.Println("New connection")

	// Upgrade the HTTP request to a WebSocket connection
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}

	// Create a new client and add it to the manager
	client := NewClient(conn, m)
	m.addClient(client)

	// Start the read and write goroutines for the client
	go client.readMessages()
	go client.writeMessages()
}

// addClient will add clients to our clientList
func (m *Manager) addClient(client *Client) {
	// Lock so we can manipulate
	m.Lock()
	defer m.Unlock()

	// Add Client
	m.clients[client] = true
}

// removeClient will remove the client and clean up
func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	// Check if Client exists, then delete it
	if _, ok := m.clients[client]; ok {
		// close connection
		client.connection.Close()
		// remove
		delete(m.clients, client)
	}
}

// checkOrigin will check origin and return true if its allowed
func checkOrigin(r *http.Request) bool {

	// Grab the request origin
	origin := r.Header.Get("Origin")

	switch origin {
	case "https://localhost:8080":
		return true
	default:
		return false
	}
}
