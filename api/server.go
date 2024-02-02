package api

import (
	"context"
	"fmt"
	"os"

	"github.com/Streamfair/streamfair_user_svc/token"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

// Server serves HTTP requests for the streamfair backend service.
type Server struct {
	config          util.Config
	store           db.Store
	localTokenMaker token.Maker
	router          *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	// localTokenMaker, err := token.NewLocalPasetoMaker(config.TokenSymmetricKey)
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to create local token maker: %v", err))
	// }

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("acctype", validAccountTypes)
	}

	server := &Server{
		config:          config,
		store:           store,
		// localTokenMaker: localTokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.GET("/readiness", server.readinessCheck)

	router.POST("/users", server.createUser)
	// router.POST("/users/login", server.loginUser)

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

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/id/:id", server.getAccountByID)
	authRoutes.GET("/accounts/id", server.handleMissingID)
	authRoutes.GET("/accounts/owner/:owner", server.getAccountByOwner)
	authRoutes.GET("/accounts/owner", server.handleMissingOwner)
	authRoutes.GET("/accounts/list", server.listAccount)
	authRoutes.PUT("/accounts/update/:id", server.updateAccount)
	authRoutes.PUT("/accounts/update", server.handleMissingID)
	authRoutes.DELETE("/accounts/delete/:id", server.deleteAccount)
	authRoutes.DELETE("/accounts/delete", server.handleMissingID)

	server.router = router
}

// StartServer starts a new HTTP server on the specified address.
func (server *Server) StartServer(address string) error {
	if err := InitializeDatabase(server.store); err != nil {
		fmt.Fprintf(os.Stderr, "database: error while initializing database: %v\n", err)
		return err
	}
	return server.router.Run(address)
}

// InitializeDatabase creates the initial fixed entries in the database.
func InitializeDatabase(store db.Store) error {
	accountTypes := util.GetAccountTypeStruct()
	arg := db.ListAccountTypesParams{
		Limit:  int32(len(accountTypes)),
		Offset: 0,
	}
	accountTypesInDB, err := store.ListAccountTypes(context.Background(), arg)
	if err != nil {
		return err
	}

	// Convert accountTypesInDB into a map for faster lookup
	accountTypesMap := make(map[int32]bool)
	for _, accountTypeInDB := range accountTypesInDB {
		accountTypesMap[accountTypeInDB.ID] = true
	}

	var errs []error
	for _, accountType := range accountTypes {
		if !accountTypesMap[int32(accountType.ID)] {
			_, err := store.CreateAccountType(context.Background(), db.CreateAccountTypeParams{
				Type:        accountType.Type,
				Permissions: accountType.Permissions,
				IsArtist:    accountType.IsArtist,
				IsProducer:  accountType.IsProducer,
				IsWriter:    accountType.IsWriter,
				IsLabel:     accountType.IsLabel,
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%q", errs)
	}

	fmt.Println("Database initialized.")
	return nil
}

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
