package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewLogParserHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

/*
 *   Parse does:
 *   1. get data from *.log files in "../logs/" directory
 *   2. parse into array of LogData type
 *   3. convert into csv files and save them in "../data/" files
 */
func (h *Handler) Parse(ctx *gin.Context) {
	if err := h.service.Parse(); err != nil {
		ctx.JSON(http.StatusInternalServerError, struct{ err string }{err: err.Error()})
		ctx.Abort()
	}

	ctx.JSON(http.StatusOK, struct{ result string }{result: "ok"})
}
