package controllers

import (
	"fmt"
	"id/projects/market-data/helper"
	"id/projects/market-data/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type analyzeController struct {
}

func NewAnalyzeController() *analyzeController {
	return &analyzeController{}
}

const defaultDate = "2006-01-02"

func (h *analyzeController) GetAnalyze(c *gin.Context) {
	var req models.AnalyzeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	start, err := time.Parse(defaultDate, req.StartDate)
	if err != nil {
		response := helper.APIResponse("Invalid start date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	end, err := time.Parse(defaultDate, req.EndDate)
	if err != nil {
		response := helper.APIResponse("Invalid end date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	stock, err := quote.NewQuoteFromYahoo(req.Symbol, start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
	if err != nil {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)

		return
	}

	// Check if the stock data is empty
	if len(stock.Close) == 0 {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// Perform technical analysis using go-talib
	closePrices := make([]float64, len(stock.Close))
	for i, price := range stock.Close {
		closePrices[i] = price
	}

	// Calculate moving averages
	ma5 := talib.Sma(closePrices, 5)
	ma10 := talib.Sma(closePrices, 10)
	ma20 := talib.Sma(closePrices, 20)
	ma50 := talib.Sma(closePrices, 50)

	// Calculate the Relative Strength Index (RSI)
	rsi := talib.Rsi(closePrices, 14)

	// Calculate the Moving Average Convergence Divergence (MACD) and its signal line
	macd, macdSignal, _ := talib.Macd(closePrices, 12, 26, 9)

	// Check if the latest close price is above or below the moving averages
	latestClose := closePrices[len(closePrices)-1]
	latestMa5 := ma5[len(ma5)-1]
	latestMa10 := ma10[len(ma10)-1]
	latestMa20 := ma20[len(ma20)-1]
	latestMa50 := ma50[len(ma50)-1]

	// Check if the MACD and its signal line are crossing
	macdCross := false
	if macd[len(macd)-1] > macdSignal[len(macdSignal)-1] && macd[len(macd)-2] < macdSignal[len(macdSignal)-2] {
		macdCross = true
	} else if macd[len(macd)-1] < macdSignal[len(macdSignal)-1] && macd[len(macd)-2] > macdSignal[len(macdSignal)-2] {
		macdCross = true
	}

	var recommendation string
	var buyTarget, sellTarget float64

	if latestClose > latestMa5 && latestClose > latestMa10 && latestClose > latestMa20 && latestClose > latestMa50 && rsi[len(rsi)-1] > 50 && macdCross {
		recommendation = "Buy"
		// Calculate the buy target as the average of the 5-day and 10-day SMAs
		buyTarget = (ma5[len(ma5)-1] + ma10[len(ma10)-1]) / 2
		// Calculate the sell target as 5% above the 20-day SMA
		sellTarget = ma20[len(ma20)-1] * 1.05
	} else if latestClose < latestMa5 && latestClose < latestMa10 && latestClose < latestMa20 && latestClose < latestMa50 && rsi[len(rsi)-1] < 50 && macdCross {
		recommendation = "Sell"
		// Calculate the buy target as 5% below the 20-day SMA
		buyTarget = ma20[len(ma20)-1] * 0.95
		// Calculate the sell target as the average of the 5-day and 10-day SMAs
		sellTarget = (ma5[len(ma5)-1] + ma10[len(ma10)-1]) / 2
	} else {
		recommendation = "Hold"
	}

	respFormatter := models.AnalyzeResponse{}
	respFormatter.Symbol = req.Symbol
	respFormatter.StartDate = start.Format(defaultDate)
	respFormatter.EndDate = end.Format(defaultDate)
	respFormatter.Recommendation = recommendation

	quoteFormatter := models.AnalyzeQuote{}
	quoteFormatter.BuyTarget = buyTarget
	quoteFormatter.SellTarget = sellTarget
	quoteFormatter.LatestClose = latestClose
	quoteFormatter.LatestMA5 = latestMa5
	quoteFormatter.LatestMA10 = latestMa10
	quoteFormatter.LatestMA20 = latestMa20
	quoteFormatter.LatestMA50 = latestMa50
	quoteFormatter.RSI = rsi[len(rsi)-1]
	quoteFormatter.MACD = macd[len(macd)-1]
	quoteFormatter.MACDSignal = macdSignal[len(macdSignal)-1]

	respFormatter.AnalyzeQuote = quoteFormatter

	response := helper.APIResponse("Analyze quote successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)
}

func (h *analyzeController) GetReceommendation(c *gin.Context) {
	var req models.RecommendationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	symbols := strings.Split(req.Symbols, ",")

	start, err := time.Parse(defaultDate, req.StartDate)
	if err != nil {
		response := helper.APIResponse("Invalid start date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	end, err := time.Parse(defaultDate, req.EndDate)
	if err != nil {
		response := helper.APIResponse("Invalid end date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// Loop over each symbol and calculate its score
	var bestSymbol string
	bestScore := -1.0
	for _, symbol := range symbols {
		stock, err := quote.NewQuoteFromYahoo(symbol, start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
		if err != nil {
			continue // Skip this symbol and move on to the next one
		}

		// Check if the stock data is empty
		if len(stock.Close) == 0 {
			continue // Skip this symbol and move on to the next one
		}

		// Perform technical analysis using go-talib
		closePrices := make([]float64, len(stock.Close))
		for i, price := range stock.Close {
			closePrices[i] = price
		}

		// Calculate moving averages
		ma5 := talib.Sma(closePrices, 5)
		ma10 := talib.Sma(closePrices, 10)
		ma20 := talib.Sma(closePrices, 20)
		ma50 := talib.Sma(closePrices, 50)

		// Calculate the RSI
		rsi := talib.Rsi(closePrices, 14)

		// Calculate the MACD and its signal line
		macd, macdSignal, _ := talib.Macd(closePrices, 12, 26, 9)

		// Check if the MACD and its signal line are crossing
		macdCross := macd[len(macd)-1] > macdSignal[len(macdSignal)-1] && macd[len(macd)-2] < macdSignal[len(macd)-2]

		// Calculate the score based on the technical analysis
		var score float64
		latestClose := closePrices[len(closePrices)-1]
		if latestClose > ma5[len(ma5)-1] && latestClose > ma10[len(ma10)-1] && latestClose > ma20[len(ma20)-1] && latestClose > ma50[len(ma50)-1] && rsi[len(rsi)-1] > 50 {
			score = 3.0 // Strong Buy
		} else if (latestClose > ma5[len(ma5)-1] && latestClose > ma10[len(ma10)-1] && latestClose > ma20[len(ma20)-1]) ||
			(rsi[len(rsi)-1] > 50 && rsi[len(rsi)-1] < 70) ||
			macdCross {
			score = 2.0 // Buy
		} else if (latestClose < ma5[len(ma5)-1] && latestClose < ma10[len(ma10)-1] && latestClose < ma20[len(ma20)-1]) ||
			(rsi[len(rsi)-1] < 50 && rsi[len(rsi)-1] > 30) {
			score = 1.0 // Sell
		} else {
			score = 0.0 // Strong Sell
		}

		// Check if this symbol has a higher score than the previous best one
		if score > bestScore {
			bestScore = score
			bestSymbol = symbol
		}
	}

	// Calculate the target prices for the best symbol
	if bestScore >= 0 {
		stock, err := quote.NewQuoteFromYahoo(bestSymbol, start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
		if err != nil {
			errors := fmt.Sprintf("Failed to retrieve stock data for %s", bestSymbol)
			response := helper.APIResponse(errors, http.StatusBadRequest, "FAILED", nil)
			c.JSON(http.StatusOK, response)
			return
		}

		// Check if the stock data is empty
		if len(stock.Close) == 0 {
			errors := fmt.Sprintf("Failed to retrieve stock data for %s", bestSymbol)
			response := helper.APIResponse(errors, http.StatusBadRequest, "FAILED", nil)
			c.JSON(http.StatusOK, response)
			return
		}

		// Perform technical analysis using go-talib
		closePrices := make([]float64, len(stock.Close))
		for i, price := range stock.Close {
			closePrices[i] = price
		}

		// Calculate the target prices to buy and sell
		latestClose := closePrices[len(closePrices)-1]
		buyTarget := latestClose * (1 + 0.01*(bestScore+1))
		sellTarget := latestClose * (1 - 0.01*(bestScore+1))

		respFormatter := models.RecommendationResponse{}
		respFormatter.BestSymbol = bestSymbol
		respFormatter.BestScore = bestScore
		respFormatter.BuyTarget = buyTarget
		respFormatter.SellTarget = sellTarget

		response := helper.APIResponse("Recommendation quote successfully", http.StatusOK, "SUCCESS", respFormatter)
		c.JSON(http.StatusOK, response)
	} else {
		response := helper.APIResponse("Failed to find any valid symbols", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)

	}
}

func (h *analyzeController) GetForecast(c *gin.Context) {
	var req models.QuoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	currentTime := time.Now()
	lastWeek := currentTime.AddDate(0, 0, -100)

	stock, err := quote.NewQuoteFromYahoo(req.Symbol, lastWeek.Format(defaultDate), currentTime.Format(defaultDate), quote.Daily, true)
	if err != nil {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// Check if the stock data is empty
	if len(stock.Close) == 0 {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// Perform technical analysis using go-talib
	closePrices := make([]float64, len(stock.Close))
	for i, price := range stock.Close {
		closePrices[i] = price
	}

	// Calculate moving averages
	ma5 := talib.Sma(closePrices, 5)
	ma10 := talib.Sma(closePrices, 10)
	ma20 := talib.Sma(closePrices, 20)
	ma50 := talib.Sma(closePrices, 50)

	// Calculate the buy target as the average of the 5-day and 10-day moving averages
	buyTarget := (ma5[len(ma5)-1] + ma10[len(ma10)-1]) / 2

	// Calculate the sell target as the average of the 20-day and 50-day moving averages
	sellTarget := (ma20[len(ma20)-1] + ma50[len(ma50)-1]) / 2

	// Calculate the RSI
	rsi := talib.Rsi(closePrices, 14)

	// Calculate the MACD
	macd, macdSignal, macdHistogram := talib.Macd(closePrices, 12, 26, 9)

	// Calculate the score based on the technical analysis
	score := 0
	if closePrices[len(closePrices)-1] > buyTarget {
		score += 1
	}
	if closePrices[len(closePrices)-1] > sellTarget {
		score += 1
	}
	if ma5[len(ma5)-1] > ma20[len(ma20)-1] && ma20[len(ma20)-1] > ma50[len(ma50)-1] {
		score += 1
	}
	if rsi[len(rsi)-1] > 50 {
		score += 1
	}
	if macdSignal[len(macdSignal)-1] > macdHistogram[len(macdHistogram)-1] {
		score += 1
	}

	// Calculate the expected price for the next day
	lastClose := closePrices[len(closePrices)-1]
	expectedPrice := lastClose * (1 + (float64(score) / 10))

	respFormatter := models.ForcestResponse{}
	respFormatter.Symbol = req.Symbol
	respFormatter.Score = score
	respFormatter.ExpectedPrice = expectedPrice

	quoteFormatter := models.ForcestQuote{}
	quoteFormatter.BuyTarget = buyTarget
	quoteFormatter.SellTarget = sellTarget
	quoteFormatter.LatestClose = lastClose
	quoteFormatter.RSI = rsi[len(rsi)-1]
	quoteFormatter.MACD = macd[len(macd)-1]
	quoteFormatter.MACDSignal = macdSignal[len(macdSignal)-1]
	quoteFormatter.MACDHistogram = macdHistogram[len(macdHistogram)-1]

	respFormatter.ForcestQuote = quoteFormatter

	response := helper.APIResponse("Forcest quote successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)
}

func (h *analyzeController) GetFundamental(c *gin.Context) {
	var req models.AnalyzeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		response := helper.APIResponse("Unable to process request", http.StatusUnprocessableEntity, "FAILED", errorMessage)
		c.JSON(http.StatusOK, response)
		return
	}

	start, err := time.Parse(defaultDate, req.StartDate)
	if err != nil {
		response := helper.APIResponse("Invalid start date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	end, err := time.Parse(defaultDate, req.EndDate)
	if err != nil {
		response := helper.APIResponse("Invalid end date format, should be YYYY-MM-DD", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	quotes, err := quote.NewQuoteFromYahoo(req.Symbol, start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
	if err != nil {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)

		return
	}

	// Calculate key financial data using go-talib
	closePrices := make([]float64, len(quotes.Close))
	for i := 0; i < len(quotes.Close); i++ {
		closePrices[i] = quotes.Close[i]
	}

	ma14 := talib.Sma(closePrices, 14)
	ema50 := talib.Ema(closePrices, 50)

	peRatio := ema50[len(ema50)-1] / ma14[len(ma14)-1]

	// Make a recommendation based on P/E ratio
	var recommendation string
	if peRatio < 15 {
		recommendation = "Buy"
	} else if peRatio > 25 {
		recommendation = "Sell"
	} else {
		recommendation = "Hold"
	}

	respFormatter := models.FundamentalResponse{}
	respFormatter.Symbol = quotes.Symbol
	respFormatter.StartDate = start.Format(defaultDate)
	respFormatter.EndDate = end.Format(defaultDate)
	respFormatter.Recommendation = recommendation

	response := helper.APIResponse("Fundamental quote successfully", http.StatusOK, "SUCCESS", respFormatter)
	c.JSON(http.StatusOK, response)
}
