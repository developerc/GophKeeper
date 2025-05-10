package entity

type CardDataDTO struct {
	Number     string `json:"number"`
	Month      string `json:"month"`
	Year       string `json:"year"`
	CardHolder string `json:"card_holder"`
}
