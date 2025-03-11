package loggraphhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	csvRepository CSVRepository
	csvParser     CSVParser
}

func NewLogGraphHandler(csvRepository CSVRepository, csvParser CSVParser) *Handler {
	return &Handler{
		csvRepository: csvRepository,
		csvParser:     csvParser,
	}
}

func (h *Handler) Graph(ctx *gin.Context) {
	data, err := h.csvRepository.GetAllCSVData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}
	logs, err := h.csvParser.Parse(data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": logs})
}
