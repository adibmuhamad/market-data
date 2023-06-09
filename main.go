package main

import (
	"id/projects/market-data/controllers"
	"id/projects/market-data/services"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	sentimentService := services.NewSentimenService()

	quoteController := controllers.NewQuoteController()
	analyzeController := controllers.NewAnalyzeController()
	sentimentController := controllers.NewNewsController(sentimentService)
	simulateController := controllers.NewSimulateController()

	router := r.Group("/api/v1")
	{
		// Quote
		router.GET("/quote", quoteController.GetQuote)
		router.GET("/index", quoteController.GetIndex)

		// Analyze
		router.GET("/analyze", analyzeController.GetAnalyze)
		router.GET("/analyze/recommendation", analyzeController.GetReceommendation)
		router.GET("/analyze/forecast", analyzeController.GetForecast)
		router.GET("/analyze/fundamental", analyzeController.GetFundamental)

		// News
		router.GET("/news/sentiment", sentimentController.GetSentiment)

		// SImulate
		router.GET("/simulate", simulateController.GetSimulate)
	}

	r.Run(":8080")
}
