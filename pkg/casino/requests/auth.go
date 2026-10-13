package requests

type Auth struct {
	Username string `json:"username"`
	Balance  int64  `json:"balance"`
}
