package minesgame

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(Game{})
}

func (game *Game) Serialise() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(game); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (game *Game) Deserialise(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(game)
}
