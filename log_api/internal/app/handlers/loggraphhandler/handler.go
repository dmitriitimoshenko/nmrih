package loggraphhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	redisCache    redisCache
	csvRepository csvRepository
	csvParser     csvParser
	graphService  graphService
}

func NewLogGraphHandler(
	redisCache redisCache,
	csvRepository csvRepository,
	csvParser csvParser,
	graphService graphService,
) *Handler {
	return &Handler{
		redisCache:    redisCache,
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

	redisCacheKey := "graph_data:" + graphType.String()
	timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second)
	defer cancel()
	cached, ok, err := h.redisCache.Get(timeoutCtx, redisCacheKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}
	if ok && cached != "" {
		ctx.JSON(http.StatusOK, gin.H{"data": cached})
		return
	}

	var result interface{}

	switch graphType {
	case enums.GraphTypes.TopTimeSpentGraphType():
		{
			result = h.graphService.TopTimeSpent(logs)
			break
		}
	case enums.GraphTypes.TopCountriesGraphType():
		{
			result = h.graphService.TopCountries(logs)
			break
		}
	case enums.GraphTypes.PlayersInfoGraphType():
		{
			result, err = h.graphService.PlayersInfo()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				ctx.Abort()
				return
			}
			break
		}
	case enums.GraphTypes.OnlineStatisticsGraphType():
		{
			result = h.graphService.OnlineStatistics(logs)
			break
		}
	default:
		{
			ctx.JSON(http.StatusOK, gin.H{"data": "none"})
			return
		}
	}

	response := gin.H{"data": result}
	responseJSONBytes, err := json.Marshal(response)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response for caching"})
		ctx.Abort()
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx.Request.Context(), time.Second)
	defer cancel()
	if err := h.redisCache.Set(timeoutCtx, redisCacheKey, string(responseJSONBytes), nil); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cache graph data"})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, response)
}
