package cli

import (
	"context"
	"fmt"

	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
)

type ArgFlag struct {
	Key      string
	Example  string
	Required bool
}

type ArgParser interface {
	Flags() []ArgFlag
	Parse(ctx context.Context, args map[string]string) (func() games.Controller, error)
}

var ParserRegistry = map[games.Type]ArgParser{}

func ParseArgs(ctx context.Context, gameType games.Type, args map[string]string) (func() games.Controller, error) {
	parser := ParserRegistry[gameType]
	if parser == nil {
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}

	return parser.Parse(ctx, args)
}
