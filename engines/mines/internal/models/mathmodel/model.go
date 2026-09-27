package mathmodel

import (
	"encoding/json"
	"os"
)

const (
	FilePath = "config/math.json"
)

type Model struct {
	MinBet    int64    `json:"min_bet"`
	MaxBet    int64    `json:"max_bet"`
	GridSize  int      `json:"grid_size"`
	HouseEdge float64  `json:"house_edge"`
	MetaData  MetaData `json:"meta_data"`
}

type MetaData struct {
	Version string `json:"version"`
}

func LoadFromFile() (*Model, error) {
	file, err := os.Open(FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	model := new(Model)
	if err := json.NewDecoder(file).Decode(model); err != nil {
		return nil, err
	}

	return model, nil
}
