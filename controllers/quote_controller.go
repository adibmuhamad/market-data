package controllers

import (
	"id/projects/market-data/helper"
	"id/projects/market-data/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/quote"
)

type quoteController struct {
}

func NewQuoteController() *quoteController {
	return &quoteController{}
}

func (h *quoteController) GetQuote(c *gin.Context) {
	var req models.QuoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	quote, err := quote.Get(req.Symbol)
	if err != nil {
		response := helper.APIResponse(err.Error(), http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// convert timestamp to time.Time object
	t := time.Unix(int64(quote.RegularMarketTime), 0)

	respFormatter := models.QuoteResponse{}
	respFormatter.Symbol = quote.Symbol
	respFormatter.Name = quote.ShortName
	respFormatter.Currency = quote.CurrencyID
	respFormatter.Price = quote.RegularMarketPrice
	respFormatter.Change = quote.RegularMarketChange
	respFormatter.PercentChange = quote.RegularMarketChangePercent
	respFormatter.Open = quote.RegularMarketOpen
	respFormatter.Close = quote.RegularMarketPreviousClose
	respFormatter.High = quote.RegularMarketDayHigh
	respFormatter.Low = quote.RegularMarketDayLow
	respFormatter.Volume = quote.RegularMarketVolume
	respFormatter.Time = t.Format(time.RFC3339)

	response := helper.APIResponse("Get quote successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)
}

func (h *quoteController) GetIndex(c *gin.Context) {
	var req models.IndexRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	var temp = "^" + req.Index

	index, err := quote.Get(temp)
	if err != nil {
		response := helper.APIResponse(err.Error(), http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// convert timestamp to time.Time object
	t := time.Unix(int64(index.RegularMarketTime), 0)

	respFormatter := models.QuoteResponse{}
	respFormatter.Symbol = req.Index
	respFormatter.Name = index.ShortName
	respFormatter.Currency = index.CurrencyID
	respFormatter.Price = index.RegularMarketPrice
	respFormatter.Change = index.RegularMarketChange
	respFormatter.PercentChange = index.RegularMarketChangePercent
	respFormatter.Open = index.RegularMarketOpen
	respFormatter.Close = index.RegularMarketPreviousClose
	respFormatter.High = index.RegularMarketDayHigh
	respFormatter.Low = index.RegularMarketDayLow
	respFormatter.Volume = index.RegularMarketVolume
	respFormatter.Time = t.Format(time.RFC3339)

	response := helper.APIResponse("Get index successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)

}
