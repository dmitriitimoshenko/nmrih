package loggraphhandler

import (
	"net/http"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/enums"
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
	graphTypeParam, ok := ctx.GetQuery("type")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid graph type"})
		ctx.Abort()
		return
	}
	graphType := enums.GraphType(graphTypeParam)
	if !graphType.IsValid() {
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

	switch graphType {
	case enums.GraphTypes.TopTimeSpentGraphType():
		{
			result := h.graphService.TopTimeSpent(logs)
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	case enums.GraphTypes.TopCountriesGraphType():
		{
			result := h.graphService.TopCountries(logs)
			ctx.JSON(http.StatusOK, gin.H{"data": result})
			return
		}
	case enums.GraphTypes.PlayersInfoGraphType():
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
