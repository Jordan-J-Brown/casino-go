package requests

type Close struct {
	SessionID string `json:"sessionID"`
	GameID    string `json:"gameID"`
}
