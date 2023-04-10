package models

type AnalyzeRequest struct {
	Symbol    string `json:"symbol"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type RecommendationRequest struct {
	Symbols   string `json:"symbols"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type AnalyzeResponse struct {
	Symbol         string       `json:"symbol"`
	StartDate      string       `json:"startDate"`
	EndDate        string       `json:"endDate"`
	Recommendation string       `json:"recommendation"`
	Explanation    string       `json:"explanation"`
	AnalyzeQuote   AnalyzeQuote `json:"analyze"`
}

type RecommendationResponse struct {
	Symbol         string  `json:"symbol"`
	Recommendation string  `json:"recommendation"`
	LatestClose    float64 `json:"latestClose"`
	TargetBuy      float64 `json:"targetBuy"`
	TargetSell     float64 `json:"targetSell"`
	Explanation    string  `json:"explanation"`
}

type ForcestResponse struct {
	Symbol         string  `json:"symbol"`
	PredictedPrice float64 `json:"predictedPrice"`
	Signal         string  `json:"signal"`
	TargetPrice    float64 `json:"targetPrice"`
}

type FundamentalResponse struct {
	Symbol         string `json:"symbol"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	Recommendation string `json:"recommendation"`
}

type AnalyzeQuote struct {
	LatestClose   float64 `json:"latestClose"`
	BuyTarget     float64 `json:"buyTarget"`
	SellTarget    float64 `json:"sellTarget"`
	StopLossPrice float64 `json:"stopLossPrice"`
	LatestMA5     float64 `json:"latestMA5"`
	LatestMA10    float64 `json:"latestMA10"`
	LatestMA20    float64 `json:"latestMA20"`
	LatestMA50    float64 `json:"latestMA50"`
	RSI           float64 `json:"rsi"`
	MACD          float64 `json:"macd"`
	MACDSignal    float64 `json:"macdSignal"`
	CCI           float64 `json:"cci"`
	ChaikinAD     float64 `json:"chaikinAD"`
}
