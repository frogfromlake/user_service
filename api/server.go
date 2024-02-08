package api

import (
	"fmt"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// Server serves HTTP requests for the streamfair user management service.
type Server struct {
	config          util.Config
	store           db.Store
	localTokenMaker token.Maker
	router          *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	localTokenMaker, err := token.NewLocalPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to create local token maker: %v", err))
	}

	server := &Server{
		config:          config,
		store:           store,
		localTokenMaker: localTokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/readiness", server.readinessCheck)

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.localTokenMaker))

	authRoutes.GET("/users/id/:id", server.getUserByID)
	authRoutes.GET("/users/id", server.handleMissingID)
	authRoutes.GET("/users/username/:username", server.getUserByUsername)
	authRoutes.GET("/users/username", server.handleMissingUsername)
	authRoutes.GET("/users/list", server.listUsers)
	authRoutes.PUT("/users/update/:id", server.updateUser)
	authRoutes.PUT("/users/update", server.handleMissingID)
	authRoutes.PUT("/users/update/email/:id", server.updateUserEmail)
	authRoutes.PUT("/users/update/email", server.handleMissingID)
	authRoutes.PUT("/users/update/username/:id", server.updateUsername)
	authRoutes.PUT("/users/update/username", server.handleMissingUsername)
	authRoutes.PUT("/users/update/password/:id", server.updateUserPassword)
	authRoutes.PUT("/users/update/password", server.handleMissingID)
	authRoutes.DELETE("/users/delete/:id", server.deleteUser)
	authRoutes.DELETE("/users/delete", server.handleMissingID)

	server.router = router
}

// StartServer starts a new HTTP server on the specified address.
func (server *Server) StartServer(address string) error {
	return server.router.Run(address)
}

// func (server *Server) RunGinServer(config util.Config, store db.Store) {
// 	err := server.StartServer(config.HttpServerAddress)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "server: error while starting server: %v\n", err)
// 	}
// }

func errorResponse(err error) gin.H {
	switch err := err.(type) {
	case *pgconn.PgError:
		// Handle pgconn.PgError
		switch err.Code {
		case "23505": // unique_violation
			return gin.H{"error": fmt.Sprintf("Unique violation error: %v: %v", err.Message, err.Hint)}
		case "23503": // foreign_key_violation
			return gin.H{"error": fmt.Sprintf("Foreign key violation error: %v: %v", err.Message, err.Hint)}
		default:
			return gin.H{"error": fmt.Sprintf("error: %v", err.Message)}
		}
	default:
		// Handle other types of errors
		return gin.H{"error": err.Error()}
	}
}
