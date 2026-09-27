package responses

type PlayResponse struct {
	State            string `json:"state"`
	PayoutMultiplier string `json:"payoutMultiplier"`
	Mines            int    `json:"mines"`
}

type NextResponse struct {
	State             string `json:"state"`
	PayoutMultiplier  string `json:"payoutMultiplier"`
	SelectedPositions []int  `json:"selectedPositions"`
}

type MetaData struct {
	Version string `json:"version"`
	Math    string `json:"math"`
}

type GameConfig struct {
	GridSize int `json:"gridSize"`
}
