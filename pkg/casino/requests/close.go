package requests

type Close struct {
	SessionHeader
	GameID string `json:"gameID"`
}
