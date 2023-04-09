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
	AnalyzeQuote   AnalyzeQuote `json:"quote"`
}

type RecommendationResponse struct {
	BestSymbol string  `json:"bestSymbol"`
	BestScore  float64 `json:"bestScore"`
	BuyTarget  float64 `json:"buyTarget"`
	SellTarget float64 `json:"sellTarget"`
}

type ForcestResponse struct {
	Symbol        string  `json:"symbol"`
	Score         int `json:"score"`
	ExpectedPrice float64 `json:"expectedPrice"`
	ForcestQuote  ForcestQuote `json:"quote"`
}

type FundamentalResponse struct {
	Symbol         string       `json:"symbol"`
	StartDate      string       `json:"startDate"`
	EndDate        string       `json:"endDate"`
	Recommendation string       `json:"recommendation"`
}

type AnalyzeQuote struct {
	BuyTarget   float64 `json:"buyTarget"`
	SellTarget  float64 `json:"sellTarget"`
	LatestClose float64 `json:"latestClose"`
	LatestMA5   float64 `json:"latestMA5"`
	LatestMA10  float64 `json:"latestMA10"`
	LatestMA20  float64 `json:"latestMA20"`
	LatestMA50  float64 `json:"latestMA50"`
	RSI         float64 `json:"rsi"`
	MACD        float64 `json:"macd"`
	MACDSignal  float64 `json:"macdSignal"`
}

type ForcestQuote struct {
	BuyTarget     float64 `json:"buyTarget"`
	SellTarget    float64 `json:"sellTarget"`
	LatestClose   float64 `json:"latestClose"`
	RSI           float64 `json:"rsi"`
	MACD          float64 `json:"macd"`
	MACDSignal    float64 `json:"macdSignal"`
	MACDHistogram float64 `json:"macdHistogram"`
}
