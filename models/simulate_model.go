package models

type SimulationRequest struct {
	Symbol    string `json:"symbol"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Cash      string `json:"cash"`
	BuyPrice  string `json:"buyPrice"`
	SellPrice string `json:"sellPrice"`
}

type SimulationResponse struct {
	Cash      float64 `json:"cash"`
	Shares    float64 `json:"shares"`
	GainLoss  float64 `json:"gainLoss"`
	TotalCost float64 `json:"totalCost"`
}
