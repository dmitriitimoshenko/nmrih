package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/a2sclient"
	a2sclientconfig "github.com/dmitriitimoshenko/nmrih/log_api/internal/app/a2sclient/config"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/handlers/loggraphhandler"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/handlers/logparserhandler"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/ipapiclient"
	ipapiclientconfig "github.com/dmitriitimoshenko/nmrih/log_api/internal/app/ipapiclient/config"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvgenerator"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvparser"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/csvrepository"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/graph"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/logparser"
	"github.com/dmitriitimoshenko/nmrih/log_api/internal/pkg/services/logrepository"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://rulat-bot.duckdns.org")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
	server := gin.Default()
	server.Use(gin.Logger())
	server.Use(gin.Recovery())
	server.Use(CORSMiddleware())

	ginMode := os.Getenv("GIN_MODE")
	log.Println("GIN Mode set to: ", ginMode)
	gin.SetMode(ginMode)

	serverPort, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Fatalln(err)
	}

	ipAPIClientConfig := ipapiclientconfig.NewIPAPIClientConfig(os.Getenv("IP_INFO_API_TOKEN"))
	ipAPIClient := ipapiclient.NewIPAPIClient(ipAPIClientConfig)
	a2sClientConfig := a2sclientconfig.NewA2SClientConfig(
		os.Getenv("SERVER_ADDR"),
		serverPort,
	)
	a2sClient, err := a2sclient.NewA2SClient(a2sClientConfig)
	if err != nil {
		log.Fatalln(err)
	}

	logRepositoryConfig := logrepository.NewConfig(os.Getenv("LOGS_STORAGE_DIRECTORY"), os.Getenv("LOGS_FILE_PATTERN"))
	logRepositoryService := logrepository.NewService(*logRepositoryConfig)
	csvGeneratorService := csvgenerator.NewCSVGenerator()
	csvRepositoryConfig := csvrepository.NewConfig(os.Getenv("CSV_STORAGE_DIRECTORY"))
	csvRepositoryService := csvrepository.NewService(*csvRepositoryConfig)
	csvParserService := csvparser.NewService()
	graphService := graph.NewService(a2sClient)

	logParserService := logparser.NewService(
		logRepositoryService,
		csvGeneratorService,
		csvRepositoryService,
		ipAPIClient,
	)

	logParserHandler := logparserhandler.NewLogParserHandler(logParserService)
	logGraphHandler := loggraphhandler.NewLogGraphHandler(csvRepositoryService, csvParserService, graphService)

	server.GET("/health-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})

	cacheStore := persistence.NewInMemoryStore(time.Minute)

	apiv1 := server.Group("/api/v1")
	apiv1.GET("/parse", cache.CachePage(cacheStore, time.Minute, logParserHandler.Parse))
	apiv1.GET("/graph", logGraphHandler.Graph)

	ports := fmt.Sprintf(":%s", os.Getenv("PORT"))
	err = server.Run(ports)
	if err != nil {
		log.Fatalf("couldn't run server: %v", err)
	}
}
