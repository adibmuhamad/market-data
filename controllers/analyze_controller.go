package controllers

import (
	"fmt"
	"id/projects/market-data/helper"
	"id/projects/market-data/models"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
	"github.com/sajari/regression"
)

type analyzeController struct {
}

func NewAnalyzeController() *analyzeController {
	return &analyzeController{}
}

const defaultDate = "2006-01-02"
const daysToLookBack = 50

type ByRecommendation []*models.RecommendationResponse

func (s ByRecommendation) Len() int {
	return len(s)
}
func (s ByRecommendation) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByRecommendation) Less(i, j int) bool {
	if s[i].Recommendation == s[j].Recommendation {
		return s[i].TargetBuy > s[j].TargetBuy
	}
	return s[i].Recommendation == "STRONG BUY" || (s[i].Recommendation == "BUY" && (s[j].Recommendation == "SELL" || s[j].Recommendation == "STRONG SELL"))
}

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

	// Check if start date and end date are within the last three months
	diff := end.Sub(start)
	if diff < (90 * 24 * time.Hour) {
		response := helper.APIResponse("Range date minumum 3 months", http.StatusBadRequest, "FAILED", nil)
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

	// Calculate CCI using a 20-day period
	cci := talib.Cci(stock.High, stock.Low, stock.Close, 20)

	// Calculate Chaikin Accumulation/Distribution line with default parameters (using high, low, close prices and volume)
	chaikinAD := talib.Ad(stock.High, stock.Low, stock.Close, stock.Volume)

	// Check if the latest close price is above or below the moving averages
	latestClose := closePrices[len(closePrices)-1]
	latestSMA5 := ma5[len(ma5)-1]
	latestSMA10 := ma10[len(ma10)-1]
	latestSMA20 := ma20[len(ma20)-1]
	latestSMA50 := ma50[len(ma50)-1]
	latestRSI := rsi[len(rsi)-1]
	latestMACD := macd[len(macd)-1]
	latestMACDSignal := macdSignal[len(macdSignal)-1]
	latestCCI := cci[len(cci)-1]
	latestChaikinAD := chaikinAD[len(chaikinAD)-1]

	// Use multiple indicators to confirm trend and momentum
	var buyCount, sellCount int
	if latestSMA5 > latestSMA10 && latestSMA10 > latestSMA20 && latestSMA20 > latestSMA50 {
		buyCount++
	}
	if latestRSI > 50 {
		buyCount++
	}
	if latestMACD > latestMACDSignal {
		buyCount++
	}
	if latestCCI > 0 {
		buyCount++
	}
	if latestChaikinAD > 0 {
		buyCount++
	}
	if latestSMA5 < latestSMA10 && latestSMA10 < latestSMA20 && latestSMA20 < latestSMA50 {
		sellCount++
	}
	if latestRSI < 50 {
		sellCount++
	}
	if latestMACD < latestMACDSignal {
		sellCount++
	}
	if latestCCI < 0 {
		sellCount++
	}
	if latestChaikinAD < 0 {
		sellCount++
	}

	// Determine recommendation based on the number of confirmations for buy and sell signals
	var recommendation string
	var targetBuy, targetSell float64

	// Default values for buy and sell targets
	targetBuy = latestClose
	targetSell = latestClose

	if buyCount == 4 && sellCount == 0 {
		recommendation = "STRONG BUY"
		targetBuy = latestClose + (latestClose-latestSMA20)*0.1 // 10% above SMA20
		targetSell = 0                                          // no sell recommendation for STRONG BUY
	} else if buyCount >= 3 && sellCount <= 1 {
		recommendation = "BUY"
		targetBuy = latestClose + (latestClose-latestSMA20)*0.05  // 5% above SMA20
		targetSell = latestClose - (latestSMA20-latestClose)*0.03 // 3% below SMA20
	} else if buyCount == 2 && sellCount == 2 {
		recommendation = "HOLD"
		targetBuy = 0  // no buy recommendation for HOLD
		targetSell = 0 // no sell recommendation for HOLD
	} else if sellCount >= 3 && buyCount <= 1 {
		recommendation = "SELL"
		targetBuy = 0                                             // no buy recommendation for SELL
		targetSell = latestClose - (latestSMA20-latestClose)*0.05 // 5% below SMA20
	} else if sellCount == 4 && buyCount == 0 {
		recommendation = "STRONG SELL"
		targetBuy = 0                                            // no buy recommendation for STRONG SELL
		targetSell = latestClose - (latestClose-latestSMA20)*0.1 // 10% below SMA20
	} else {
		recommendation = "NO RECOMMENDATION"
	}

	// Calculate stop-loss order price based on the most recent closing price
	stopLoss := latestClose * 0.95 // 5% below closing price

	// Explain the recommendation based on the number of confirmations for buy and sell signals
	var explanation string
	if buyCount == 4 && sellCount == 0 {
		explanation = "The stock is showing very strong buy signals from all indicators, and there are no sell signals. This is a good opportunity to buy the stock with a target price of " + fmt.Sprintf("%.2f", targetBuy) + "."
	} else if buyCount >= 3 && sellCount <= 1 {
		explanation = "The stock is showing strong buy signals from most indicators, and there are very few sell signals. This is a good opportunity to buy the stock with a target price of " + fmt.Sprintf("%.2f", targetBuy) + "."
	} else if buyCount == 2 && sellCount == 2 {
		explanation = "The stock is showing mixed signals from the indicators, and there are no clear buy or sell signals. It may be best to hold off on buying or selling the stock at this time."
	} else if sellCount >= 3 && buyCount <= 1 {
		explanation = "The stock is showing strong sell signals from most indicators, and there are very few buy signals. It may be best to sell the stock with a target price of " + fmt.Sprintf("%.2f", targetSell) + "."
	} else if sellCount == 4 && buyCount == 0 {
		explanation = "The stock is showing very strong sell signals from all indicators, and there are no buy signals. It may be best to sell the stock with a target price of " + fmt.Sprintf("%.2f", targetSell) + "."
	} else {
		explanation = "There is no clear recommendation for this stock based on the current indicators. It may be best to hold off on buying or selling the stock at this time."
	}

	respFormatter := models.AnalyzeResponse{}
	respFormatter.Symbol = req.Symbol
	respFormatter.StartDate = start.Format(defaultDate)
	respFormatter.EndDate = end.Format(defaultDate)
	respFormatter.Recommendation = recommendation
	respFormatter.Explanation = explanation

	quoteFormatter := models.AnalyzeQuote{}
	quoteFormatter.BuyTarget = targetBuy
	quoteFormatter.SellTarget = targetSell
	quoteFormatter.StopLossPrice = stopLoss
	quoteFormatter.LatestClose = latestClose
	quoteFormatter.LatestMA5 = latestSMA5
	quoteFormatter.LatestMA10 = latestSMA10
	quoteFormatter.LatestMA20 = latestSMA20
	quoteFormatter.LatestMA50 = latestSMA50
	quoteFormatter.RSI = rsi[len(rsi)-1]
	quoteFormatter.MACD = macd[len(macd)-1]
	quoteFormatter.MACDSignal = macdSignal[len(macdSignal)-1]
	quoteFormatter.CCI = cci[len(cci)-1]
	quoteFormatter.ChaikinAD = chaikinAD[len(chaikinAD)-1]

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

	// Check if start date and end date are within the last three months
	diff := end.Sub(start)
	if diff < (90 * 24 * time.Hour) {
		response := helper.APIResponse("Range date minumum 3 months", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	var stocks []*models.RecommendationResponse
	for _, symbol := range symbols {
		// Retrieve stock data
		stock, err := quote.NewQuoteFromYahoo(strings.TrimSpace(symbol), start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
		if err != nil {
			continue
		}

		// Check if the stock data is empty
		if len(stock.Close) == 0 {
			continue
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

		// Calculate CCI using a 20-day period
		cci := talib.Cci(stock.High, stock.Low, stock.Close, 20)

		// Calculate Chaikin Accumulation/Distribution line with default parameters (using high, low, close prices and volume)
		chaikinAD := talib.Ad(stock.High, stock.Low, stock.Close, stock.Volume)

		// Check if the latest close price is above or below the moving averages
		latestClose := closePrices[len(closePrices)-1]
		latestSMA5 := ma5[len(ma5)-1]
		latestSMA10 := ma10[len(ma10)-1]
		latestSMA20 := ma20[len(ma20)-1]
		latestSMA50 := ma50[len(ma50)-1]
		latestRSI := rsi[len(rsi)-1]
		latestMACD := macd[len(macd)-1]
		latestMACDSignal := macdSignal[len(macdSignal)-1]
		latestCCI := cci[len(cci)-1]
		latestChaikinAD := chaikinAD[len(chaikinAD)-1]

		// Use multiple indicators to confirm trend and momentum
		var buyCount, sellCount int
		if latestSMA5 > latestSMA10 && latestSMA10 > latestSMA20 && latestSMA20 > latestSMA50 {
			buyCount++
		}
		if latestRSI > 50 {
			buyCount++
		}
		if latestMACD > latestMACDSignal {
			buyCount++
		}
		if latestCCI > 0 {
			buyCount++
		}
		if latestChaikinAD > 0 {
			buyCount++
		}
		if latestSMA5 < latestSMA10 && latestSMA10 < latestSMA20 && latestSMA20 < latestSMA50 {
			sellCount++
		}
		if latestRSI < 50 {
			sellCount++
		}
		if latestMACD < latestMACDSignal {
			sellCount++
		}
		if latestCCI < 0 {
			sellCount++
		}
		if latestChaikinAD < 0 {
			sellCount++
		}

		// Determine recommendation based on the number of confirmations for buy and sell signals
		var recommendation string
		var targetBuy, targetSell float64

		// Default values for buy and sell targets
		targetBuy = latestClose
		targetSell = latestClose

		if buyCount == 4 && sellCount == 0 {
			recommendation = "STRONG BUY"
			targetBuy = latestClose + (latestClose-latestSMA20)*0.1 // 10% above SMA20
			targetSell = 0                                          // no sell recommendation for STRONG BUY
		} else if buyCount >= 3 && sellCount <= 1 {
			recommendation = "BUY"
			targetBuy = latestClose + (latestClose-latestSMA20)*0.05  // 5% above SMA20
			targetSell = latestClose - (latestSMA20-latestClose)*0.03 // 3% below SMA20
		} else if buyCount == 2 && sellCount == 2 {
			recommendation = "HOLD"
			targetBuy = 0  // no buy recommendation for HOLD
			targetSell = 0 // no sell recommendation for HOLD
		} else if sellCount >= 3 && buyCount <= 1 {
			recommendation = "SELL"
			targetBuy = 0                                             // no buy recommendation for SELL
			targetSell = latestClose - (latestSMA20-latestClose)*0.05 // 5% below SMA20
		} else if sellCount == 4 && buyCount == 0 {
			recommendation = "STRONG SELL"
			targetBuy = 0                                            // no buy recommendation for STRONG SELL
			targetSell = latestClose - (latestClose-latestSMA20)*0.1 // 10% below SMA20
		} else {
			recommendation = "NO RECOMMENDATION"
		}

		// Explain the recommendation based on the number of confirmations for buy and sell signals
		var explanation string
		if buyCount == 4 && sellCount == 0 {
			explanation = "The stock is showing very strong buy signals from all indicators, and there are no sell signals. This is a good opportunity to buy the stock with a target price of " + fmt.Sprintf("%.2f", targetBuy) + "."
		} else if buyCount >= 3 && sellCount <= 1 {
			explanation = "The stock is showing strong buy signals from most indicators, and there are very few sell signals. This is a good opportunity to buy the stock with a target price of " + fmt.Sprintf("%.2f", targetBuy) + "."
		} else if buyCount == 2 && sellCount == 2 {
			explanation = "The stock is showing mixed signals from the indicators, and there are no clear buy or sell signals. It may be best to hold off on buying or selling the stock at this time."
		} else if sellCount >= 3 && buyCount <= 1 {
			explanation = "The stock is showing strong sell signals from most indicators, and there are very few buy signals. It may be best to sell the stock with a target price of " + fmt.Sprintf("%.2f", targetSell) + "."
		} else if sellCount == 4 && buyCount == 0 {
			explanation = "The stock is showing very strong sell signals from all indicators, and there are no buy signals. It may be best to sell the stock with a target price of " + fmt.Sprintf("%.2f", targetSell) + "."
		} else {
			explanation = "There is no clear recommendation for this stock based on the current indicators. It may be best to hold off on buying or selling the stock at this time."
		}

		// Create Stock object and add to stocks slice
		temp := &models.RecommendationResponse{
			Symbol:         symbol,
			Recommendation: recommendation,
			LatestClose:    closePrices[len(closePrices)-1],
			TargetBuy:      targetBuy,
			TargetSell:     targetSell,
			Explanation:    explanation,
		}
		stocks = append(stocks, temp)
	}

	// Sort stocks by recommendation and target buy price
	sort.Sort(ByRecommendation(stocks))

	response := helper.APIResponse("Recommendation stocks successfully", http.StatusOK, "SUCCESS", stocks)
	c.JSON(http.StatusOK, response)

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

	end := time.Now()
	start := end.AddDate(0, 0, -(daysToLookBack * 2))

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

	if len(closePrices) < daysToLookBack {
		response := helper.APIResponse("Not enough data to calculate future price", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
	}

	// Create a regression model
	model := new(regression.Regression)
	model.SetObserved("Stock Price")
	model.SetVar(0, "Day")

	// Add data points to the model
	for i, quote := range closePrices {
		model.Train(regression.DataPoint(float64(quote), []float64{float64(i)}))
	}

	// Fit the model
	err = model.Run()
	if err != nil {
		response := helper.APIResponse("Error fitting ARIMA model", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
	}

	// Predict the stock price for the next day
	predictedPrice, err := model.Predict([]float64{float64(len(closePrices))})
	if err != nil {
		response := helper.APIResponse("Error making ARIMA prediction", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
	}

	shortTermEMA := exponentialMovingAverage(closePrices[len(closePrices)-10:], 2.0/float64(10+1))
	longTermEMA := exponentialMovingAverage(closePrices[len(closePrices)-daysToLookBack:], 2.0/float64(daysToLookBack+1))

	var signal string
	if shortTermEMA > longTermEMA {
		signal = "BUY"
	} else {
		signal = "SELL"
	}

	const profitTargetPercentage = 5.0
	var targetPrice float64
	if signal == "BUY" {
		targetPrice = predictedPrice * (1 + profitTargetPercentage/100)
	} else {
		targetPrice = predictedPrice * (1 - profitTargetPercentage/100)
	}

	respFormatter := models.ForcestResponse{}
	respFormatter.Symbol = req.Symbol
	respFormatter.PredictedPrice = predictedPrice
	respFormatter.Signal = signal
	respFormatter.TargetPrice = targetPrice

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

func exponentialMovingAverage(closePrices []float64, alpha float64) float64 {
	ema := closePrices[0]

	for i := 1; i < len(closePrices); i++ {
		ema = alpha*closePrices[i] + (1-alpha)*ema
	}
	return ema
}
