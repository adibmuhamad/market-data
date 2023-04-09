package controllers

import (
	"id/projects/market-data/helper"
	"id/projects/market-data/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

type simulateController struct {
}

func NewSimulateController() *simulateController {
	return &simulateController{}
}

func (h *simulateController) GetSimulate(c *gin.Context) {
	var req models.SimulationRequest

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

	cash, err := strconv.ParseFloat(req.Cash, 64)
	if err != nil {
		response := helper.APIResponse("Invalid Cash", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// threshold value that represents the minimum price at which we should buy shares
	buyPriceThreshold, err := strconv.ParseFloat(req.BuyPrice, 64)
	if err != nil {
		response := helper.APIResponse("Invalid buy price", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// threshold value that represents the maximum price at which we should sell shares
	sellPriceThreshold, err := strconv.ParseFloat(req.SellPrice, 64)
	if err != nil {
		response := helper.APIResponse("Invalid sell price", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	quote, err := quote.NewQuoteFromYahoo(req.Symbol, start.Format(defaultDate), end.Format(defaultDate), quote.Daily, true)
	if err != nil {
		response := helper.APIResponse(err.Error(), http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)

		return
	}

	// Check if the stock data is empty
	if len(quote.Close) == 0 {
		response := helper.APIResponse("Failed to retrieve stock data", http.StatusBadRequest, "FAILED", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	// Perform technical analysis using go-talib
	closePrices := make([]float64, len(quote.Close))
	for i, price := range quote.Close {
		closePrices[i] = price
	}

	result, err := simulateTrading(closePrices, cash, buyPriceThreshold, sellPriceThreshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := helper.APIResponse("Simulate quote successfully", http.StatusOK, "SUCCESS", result)
	c.JSON(http.StatusOK, response)
}

func simulateTrading(closePrices []float64, initialCash float64, buyThreshold float64, sellThreshold float64) (models.SimulationResponse, error) {
	// window size is the number of days used to calculate the historical average price, which is used to determine whether to buy or sell shares
	// Compute the historical average price, and RSI
	windowSize := 30
	averagePrices := talib.Sma(closePrices, windowSize)
	rsi := talib.Rsi(closePrices, 14)

	// Use momentum-based strategy with RSI, stop-loss, and trailing
	cash := initialCash
	shares := 0.0
	totalCost := 0.0
	stopLoss := 0.95    // sell if price drops 5% below purchase price
	trailingStop := 0.1 // sell if price drops 10% below highest price since purchase
	highestPrice := 0.0
	for i, stockPrice := range closePrices {
		if i < windowSize {
			continue
		}
		momentum := stockPrice / averagePrices[i-windowSize]
		if momentum > sellThreshold && shares > 0 {
			lotsToSell := shares
			cash += lotsToSell * stockPrice
			shares -= lotsToSell
			if shares > 0 {
				totalCost -= lotsToSell * totalCost / shares
			} else {
				totalCost = 0.0
			}
		} else if momentum < buyThreshold && cash > 0 {
			// Only buy if RSI is below 30 (oversold)
			if rsi[i] < 30 {
				lotsToBuy := cash / stockPrice
				shares += lotsToBuy
				totalCost += lotsToBuy * stockPrice
				cash -= lotsToBuy * stockPrice
				highestPrice = stockPrice
			}
		} else if shares > 0 {
			// Check for stop-loss and trailing stop orders
			if stockPrice < totalCost*stopLoss {
				lotsToSell := shares
				cash += lotsToSell * stockPrice
				shares = 0
				totalCost = 0.0
			} else if stockPrice > highestPrice {
				highestPrice = stockPrice
			} else if stockPrice < highestPrice*(1-trailingStop) {
				lotsToSell := shares
				cash += lotsToSell * stockPrice
				shares = 0
				totalCost = 0.0
			}
		}
	}

	// gainLoss represents the total amount of profit or loss made by the trading algorithm during the simulation
	// totalCost represents the total cost of all the shares bought during the simulation.
	// It is possible for the cash variable to be equal to the gainLoss variable at the end of the simulation. This can happen if all the shares that were bought during the simulation have been sold, and there is no remaining cash or shares at the end of the simulation.
	// When this happens, the gainLoss variable represents the total profit or loss of the trading strategy, which takes into account both the cash and the shares held during the simulation. If the gainLoss variable is equal to the cash variable, it means that all the available funds have been fully utilized during the simulation, and there is no remaining profit or loss to be realized.
	gainLoss := cash + shares*closePrices[len(closePrices)-1] - totalCost

	result := models.SimulationResponse{
		Cash:      cash,
		Shares:    shares,
		GainLoss:  gainLoss,
		TotalCost: totalCost,
	}
	return result, nil

}
