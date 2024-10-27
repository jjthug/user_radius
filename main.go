package main

import (
	"context"
	"fmt"
	"homo_hunter_backend/api"
	"homo_hunter_backend/util"
	"homo_hunter_backend/ws_worker"
	"log"
	"net/http"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		fmt.Println("cannot load config")
		log.Fatal("cannot load config")
	}

	fmt.Println("loaded config")

	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI("abcd"))
	if err != nil {
		log.Fatal("connection error", err)
	}
	defer mongoClient.Disconnect(context.Background())

	err = mongoClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("ping failed")
	}

	log.Print("mongo connected")

	setupAPI(ctx, config.RedisServer)
	// Serve on port :8080, fudge yeah hardcoded port
	err = http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// setupAPI will start all Routes and their Handlers
func setupAPI(ctx context.Context, redisServer string) {
	redisOpt := asynq.RedisClientOpt{
		Addr: redisServer,
	}
	manager := ws_worker.NewManager(ctx, redisOpt)
	// Serve the ./frontend directory at Route /
	// http.Handle("/", http.FileServer(http.Dir("./frontend")))
	// http.HandleFunc("/ws", manager.ServeWS)
	// http.HandleFunc("/login", manager.LoginHandler)
	// http.HandleFunc("/getNearbyUsers")

	server, err := api.NewServer(config, manager)

	if err != nil {
		log.Fatal("error creating server", err)
	}
	err = server.RunHTTPServer(config.ServerAddress)

	if err != nil {
		log.Fatal("cannot start server", err)
	}

}
