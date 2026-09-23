package models

type RoundResult struct {
	Closed   bool  `json:"closed"`
	Rolled   int64 `json:"rolled"`
	Winnings int64 `json:"winnings"`
}
