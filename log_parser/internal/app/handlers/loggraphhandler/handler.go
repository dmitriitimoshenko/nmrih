package loggraphhandler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	csvRepository CSVRepository
	csvParser     CSVParser
	graphService  GraphService
}

func NewLogGraphHandler(csvRepository CSVRepository, csvParser CSVParser, graphService GraphService) *Handler {
	return &Handler{
		csvRepository: csvRepository,
		csvParser:     csvParser,
		graphService:  graphService,
	}
}

func (h *Handler) Graph(ctx *gin.Context) {
	graphType, ok := ctx.GetQuery("type")
	if !ok || (graphType != "top-time-spent" && graphType != "top-country") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid graph type"})
		ctx.Abort()
		return
	}
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

	log.Printf("[Graph Handler] Logs: %+v", logs)

	switch graphType {
	case "top-time-spent":
		{
			result := h.graphService.TopTimeSpent(logs)
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	case "top-country":
		{
			ctx.JSON(http.StatusOK, gin.H{"data": "bad type"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"data": "none"})
}
