package games

import "fmt"

type GameState uint8

const (
	GameStateReady GameState = iota
	GameStatePlay
	GameStateNext
	GameStateClosed
)

func (state GameState) String() string {
	switch state {
	case GameStateClosed:
		return "GameStateClosed"
	case GameStateNext:
		return "GameStateNext"
	case GameStatePlay:
		return "GameStatePlay"
	case GameStateReady:
		return "GameStateReady"
	default:
		panic(fmt.Sprintf("unexpected session.GameState: %#v", state))
	}
}
