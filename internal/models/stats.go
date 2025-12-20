package models

type Stats struct {
	AverageRating    string `json:"average_rating"`
	TotalEvaluations string `json:"total_evaluations"`
	TotalUsers       string `json:"total_users"`
}