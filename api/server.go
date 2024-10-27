package api

import (
	"fmt"
	db "server/db/sqlc"
	"server/token"
	"server/util"

	"homo_hunter_backend/ws_worker"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, manager *ws_worker.Manager) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker([]byte(config.TokenSymmetric))
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
	}

	router := gin.Default()

	router.POST("/user", server.CreateNewUser)
	router.GET("/login", server.Login)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.GET("/user", server.GetUser)
	authRoutes.GET("/ws", manager.ServeWS)
	authRoutes.GET("/getNearbyUsers")

	server.router = router
	return server, nil
}

func (server *Server) RunHTTPServer(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
