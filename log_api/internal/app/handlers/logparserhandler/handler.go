package logparserhandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	redisCache redisCache
	service    service
}

func NewLogParserHandler(
	redisCache redisCache,
	service service,
) *Handler {
	return &Handler{
		redisCache: redisCache,
		service:    service,
	}
}

/*
 *   Parse does:
 *   1. get data from *.log files in "../logs/" directory
 *   2. parse into array of LogData type
 *   3. convert into csv files and save them in "../data/" files
 *   4. flush redis cache
 */
func (h *Handler) Parse(ctx *gin.Context) {
	requestTimeStamp := time.Now()

	if err := h.service.Parse(requestTimeStamp); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}

	if err := h.redisCache.FlushAll(ctx); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logs have been parsed successfully"})
}
