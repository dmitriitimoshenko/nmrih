package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/app/handlers"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/app/ipapiclient"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/csvgenerator"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/csvrepository"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/logparser"
	"github.com/dmitriitimoshenko/nmrih/log_parser/internal/pkg/services/logrepository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}

	server := gin.Default()
	server.Use(gin.Logger())
	server.Use(gin.Recovery())
	gin.SetMode(os.Getenv("GIN_MODE"))

	logRepositoryService := logrepository.NewService()
	csvGeneratorService := csvgenerator.NewCSVGenerator()
	csvRepositoryService := csvrepository.NewService()
	ipAPIClient := ipapiclient.NewIPAPIClient()

	logParserService := logparser.NewService(
		logRepositoryService,
		csvGeneratorService,
		csvRepositoryService,
		ipAPIClient,
	)

	logparserhandler := handlers.NewLogParserHandler(logParserService)

	apiv1 := server.Group("/api/v1")
	apiv1.GET("/parse", logparserhandler.Parse)

	ports := fmt.Sprintf(":%s", os.Getenv("PORT"))
	err = server.Run(ports)
	if err != nil {
		log.Fatalf("couldn't run server: %v", err)
	}
}
