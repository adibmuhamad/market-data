package models

type NewsArticle struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type NewsResponse struct {
	Articles []NewsArticle `json:"articles"`
}

type NewsAPIParams struct {
	SearchTerm string `url:"searchTerm"`
}

type SentimentResponse struct {
	Symbol    string `json:"symbol"`
	Sentiment string `json:"sentiment"`
}
