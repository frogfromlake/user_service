package api

import (
	db "github.com/frogfromlake/user_service/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for the streamfair backend service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("acctype", validAccountTypes)
	}

	router.GET("/readiness", server.readinessCheck)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccountByID)
	router.GET("/accounts/name/:username", server.getAccountByUsername)
	router.GET("/accounts/name", server.handleMissingUsername)
	router.GET("/accounts/params", server.getAccountbyAllParams)
	router.GET("/accounts", server.listAccount)
	router.PUT("/accounts/:id", server.updateAccount)
	router.PUT("/accounts/password/:id", server.updateAccountPassword)
	router.DELETE("/accounts/:id", server.deleteAccount)

	router.POST("/songs", server.addSong)
	router.PUT("/songs/:id", server.updateSong)
	router.DELETE("/songs/:id", server.deleteSong)
	server.router = router
	return server
}

// StartServer starts a new HTTP server on the specified address.
func (server *Server) StartServer(address string) error {
	// if err := InitializeDatabase(server.store); err != nil {
	// 	fmt.Fprintf(os.Stderr, "database: error while initializing database: %v\n", err)
	// }
	return server.router.Run(address)
}

// InitializeDatabase creates the initial fixed entries in the database.
// func InitializeDatabase(store db.Store) error {
// 	accountTypes := util.GetAccountTypeStruct()
// 	arg := db.ListAccountTypesParams{
// 		Limit:  int32(len(accountTypes)),
// 		Offset: 0,
// 	}
// 	accountTypesInDB, err := store.ListAccountTypes(context.Background(), arg)
// 	if err != nil {
// 		return err
// 	}

// 	// Convert accountTypesInDB into a map for faster lookup
// 	accountTypesMap := make(map[int64]bool)
// 	for _, accountTypeInDB := range accountTypesInDB {
// 		accountTypesMap[accountTypeInDB.ID] = true
// 	}

// 	var errs []error
// 	for _, accountType := range accountTypes {
// 		if !accountTypesMap[accountType.ID] {
// 			_, err := store.CreateAccountType(context.Background(), db.CreateAccountTypeParams{
// 				Description: accountType.Description,
// 				Permissions: accountType.Permissions,
// 				IsArtist:    accountType.IsArtist,
// 				IsProducer:  accountType.IsProducer,
// 				IsWriter:    accountType.IsWriter,
// 				IsLabel:     accountType.IsLabel,
// 			})
// 			if err != nil {
// 				errs = append(errs, err)
// 			}
// 		}
// 	}

// 	if len(errs) > 0 {
// 		return fmt.Errorf("%q", errs)
// 	}

// 	fmt.Println("Database initialized.")
// 	return nil
// }

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
