package requests

type PlayRequest struct {
	Mines int `json:"mines"`
}

type NextRequest struct {
	Cashout bool `json:"cashout"`
	Index   int  `json:"index"`
}
