package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) handleMissingID(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("missing 'id' in request uri")))
}

func (server *Server) handleMissingUsername(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("missing 'username' in request uri")))
}
