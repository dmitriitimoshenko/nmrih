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

func NewLogGraphHandler(
	csvRepository CSVRepository,
	csvParser CSVParser,
	graphService GraphService,
) *Handler {
	return &Handler{
		csvRepository: csvRepository,
		csvParser:     csvParser,
		graphService:  graphService,
	}
}

func (h *Handler) Graph(ctx *gin.Context) {
	graphType, ok := ctx.GetQuery("type")
	if !ok || (graphType != "top-time-spent" && graphType != "top-country" && graphType != "players-info") {
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

	log.Println("[GraphHandler] Logs:")
	for i, logE := range logs {
		if logE != nil {
			log.Printf("[Graph Handler] [%d]: %v", i+1, *logE)
		} else {
			log.Printf("[Graph Handler] [%d]: {{{ EMPTY }}}", i+1)
		}
	}

	switch graphType {
	case "top-time-spent":
		{
			result := h.graphService.TopTimeSpent(logs)
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	case "top-country":
		{
			result := h.graphService.TopCountries(logs)
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	case "players-info":
		{
			result, err := h.graphService.PlayersInfo()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				ctx.Abort()
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"data": "none"})
}
