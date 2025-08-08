package loggraphhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/dto"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/enums"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	redisCache    redisCache
	csvRepository csvRepository
	csvParser     csvParser
	graphService  graphService
	defaultTTL    time.Duration
}

func NewLogGraphHandler(
	redisCache redisCache,
	csvRepository csvRepository,
	csvParser csvParser,
	graphService graphService,
) *Handler {
	logGraphHandlerCacheTTLMinutes, err := strconv.Atoi(os.Getenv("LOG_GRAPH_HANDLER_CACHE_TTL_MINUTES"))
	if err != nil || logGraphHandlerCacheTTLMinutes <= 0 {
		fmt.Println("LOG_GRAPH_HANDLER_CACHE_TTL_MINUTES not set or invalid, using default value of 5")
		logGraphHandlerCacheTTLMinutes = 5
	}
	logGraphHandlerCacheTTL := time.Duration(logGraphHandlerCacheTTLMinutes) * time.Minute

	return &Handler{
		redisCache:    redisCache,
		csvRepository: csvRepository,
		csvParser:     csvParser,
		graphService:  graphService,
		defaultTTL:    logGraphHandlerCacheTTL,
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

	response, ok := h.getResponseByGraphType(graphType, logs)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, response)
		ctx.Abort()
		return
	}

	responseJSONBytes, err := json.Marshal(response)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response for caching"})
		ctx.Abort()
		return
	}

	timeoutCtx, cancel = context.WithTimeout(ctx.Request.Context(), time.Second)
	defer cancel()
	if err := h.redisCache.Set(timeoutCtx, redisCacheKey, string(responseJSONBytes), &h.defaultTTL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cache graph data"})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) getResponseByGraphType(graphType enums.GraphType, logs []*dto.LogData) (gin.H, bool) {
	switch graphType {
	case enums.GraphTypes.TopTimeSpentGraphType():
		{
			return gin.H{"data": h.graphService.TopTimeSpent(logs)}, true
		}
	case enums.GraphTypes.TopCountriesGraphType():
		{
			return gin.H{"data": h.graphService.TopCountries(logs)}, true
		}
	case enums.GraphTypes.PlayersInfoGraphType():
		{
			result, err := h.graphService.PlayersInfo()
			if err != nil {
				return gin.H{"error": err.Error()}, false
			}
			return gin.H{"data": result}, true
		}
	case enums.GraphTypes.OnlineStatisticsGraphType():
		{
			return gin.H{"data": h.graphService.OnlineStatistics(logs)}, true
		}
	default:
		{
			return gin.H{"data": "none"}, true
		}
	}
}
