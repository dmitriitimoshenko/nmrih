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
	cacheTimeout  time.Duration
}

func NewLogGraphHandler(
	redisCache redisCache,
	csvRepository csvRepository,
	csvParser csvParser,
	graphService graphService,
) *Handler {
	logGraphHandlerCacheTTLMinutes, err := strconv.Atoi(os.Getenv("LOG_GRAPH_HANDLER_CACHE_TTL_MINUTES"))
	if err != nil || logGraphHandlerCacheTTLMinutes <= 0 {
		fmt.Println("LOG_GRAPH_HANDLER_CACHE_TTL_MINUTES not set or invalid, using default value of 5: " + err.Error())
		logGraphHandlerCacheTTLMinutes = 5
	}
	logGraphHandlerCacheTTL := time.Duration(logGraphHandlerCacheTTLMinutes) * time.Minute

	cacheTimeoutSeconds, err := strconv.Atoi(os.Getenv("LOG_GRAPH_HANDLER_CACHE_TIMEOUT_SECONDS"))
	if err != nil || cacheTimeoutSeconds <= 0 {
		fmt.Println("LOG_GRAPH_HANDLER_CACHE_TIMEOUT_SECONDS not set or invalid, using default value of 10: " + err.Error())
		cacheTimeoutSeconds = 10
	}
	cacheTimeout := time.Duration(cacheTimeoutSeconds) * time.Second

	return &Handler{
		redisCache:    redisCache,
		csvRepository: csvRepository,
		csvParser:     csvParser,
		graphService:  graphService,
		defaultTTL:    logGraphHandlerCacheTTL,
		cacheTimeout:  cacheTimeout,
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

	cached, err := h.getCacheIfApplicable(ctx, graphType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}
	if cached != nil {
		var response gin.H
		if err := json.Unmarshal([]byte(*cached), &response); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}
		ctx.JSON(http.StatusOK, response)
		return
	}

	response, ok := h.getResponseByGraphType(graphType, logs)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, response)
		ctx.Abort()
		return
	}

	if err := h.saveCacheIfApplicable(ctx, graphType, response); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) getCacheIfApplicable(ctx context.Context, graphType enums.GraphType) (*string, error) {
	if !graphType.CanCache() {
		return nil, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.cacheTimeout)
	defer cancel()
	cached, ok, err := h.redisCache.Get(timeoutCtx, "graph_data:"+graphType.String())
	if err != nil {
		fmt.Println("Error getting from cache:", err)
		return nil, err
	}
	if ok && cached != "" {
		return &cached, nil
	}

	return nil, nil
}

func (h *Handler) saveCacheIfApplicable(ctx context.Context, graphType enums.GraphType, response gin.H) error {
	if !graphType.CanCache() {
		return nil
	}

	responseJSONBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.cacheTimeout)
	defer cancel()
	if err := h.redisCache.Set(
		timeoutCtx,
		"graph_data:"+graphType.String(),
		string(responseJSONBytes),
		&h.defaultTTL,
	); err != nil {
		return fmt.Errorf("failed to save cached graph data: %w", err)
	}

	return nil
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
