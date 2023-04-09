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

	router := r.Group("/api/v1")
	{
		// Quote
		router.GET("/quote", quoteController.GetQuote)

		// Analyze
		router.GET("/analyze", analyzeController.GetAnalyze)
		router.GET("/analyze/recommendation", analyzeController.GetReceommendation)
		router.GET("/analyze/forecast", analyzeController.GetForecast)
		router.GET("/analyze/fundamental", analyzeController.GetFundamental)

		// News
		router.GET("/news/sentiment", sentimentController.GetSentiment)
	}

	r.Run(":8080")
}
