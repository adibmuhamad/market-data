package services

import "github.com/jbrukh/bayesian"

type sentimentService struct {
	Model *bayesian.Classifier
}

type SentimenService interface {
	SentimentAnalysis(text string) int
}

func NewSentimenService() *sentimentService {
	model := bayesian.NewClassifier("negative", "positive")
	return &sentimentService{Model: model}
}

func (s *sentimentService) SentimentAnalysis(text string) int {
	scores, _, _ := s.Model.LogScores([]string{text})
	return int(scores[1] - scores[0])
}
