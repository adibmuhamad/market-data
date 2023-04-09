package models

type QuoteRequest struct {
	Symbol string `json:"symbol"`
}

type QuoteResponse struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Currency      string  `json:"currency"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	PercentChange float64 `json:"percentChange"`
	Open          float64 `json:"open"`
	Close         float64 `json:"close"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume           int `json:"volume"`
	Time          string  `json:"time"`
}
