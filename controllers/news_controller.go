package controllers

import (
	"id/projects/market-data/helper"
	"id/projects/market-data/models"
	"id/projects/market-data/services"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/gin-gonic/gin"
)

type newsController struct {
	sentimentService services.SentimenService
}

func NewNewsController(sentimentService services.SentimenService) *newsController {
	return &newsController{sentimentService}
}

func (h *newsController) GetSentiment(c *gin.Context) {
	var req models.QuoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	yahooAPI := sling.New().Base("https://finance.yahoo.com/")
	newsAPIPath := "_finance_api/resource/searchassist"
	newsAPIParams := &models.NewsAPIParams{
		SearchTerm: req.Symbol,
	}
	newsAPIResponse := new(models.NewsResponse)

	_, err := yahooAPI.Get(newsAPIPath).
		QueryStruct(newsAPIParams).
		ReceiveSuccess(newsAPIResponse)
	if err != nil {
		response := helper.APIResponse("Failed to fetch news articles", http.StatusUnprocessableEntity, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	var overallSentiment int
	for _, article := range newsAPIResponse.Articles {
		articleSentiment := h.sentimentService.SentimentAnalysis(article.Title + " " + article.Body)
		overallSentiment += articleSentiment
	}

	var sentiment string
	if overallSentiment > 0 {
		sentiment = "positive"
	} else if overallSentiment < 0 {
		sentiment = "negative"
	} else {
		sentiment = "neutral"
	}

	respFormatter := models.SentimentResponse{}
	respFormatter.Symbol = req.Symbol
	respFormatter.Sentiment = sentiment

	response := helper.APIResponse("Sentiment quote successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)
}
