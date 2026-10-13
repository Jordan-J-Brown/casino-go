package responses

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message,omitzero"`
}

func (err *Error) Error() string {
	if err.Message != "" {
		return err.Type + ": " + err.Message
	}

	return err.Type
}
