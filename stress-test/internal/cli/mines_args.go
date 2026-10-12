package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Jordan-J-Brown/casino-go/pkg/contextx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games/minescontroller"
)

func init() {
	ParserRegistry[games.TypeMines] = &MinesArgParser{}
}

type MinesArgs struct {
	Mines     int
	Positions []int
	Seed      *string
}

type MinesArgParser struct{}

func (parser *MinesArgParser) Flags() []ArgFlag {
	return []ArgFlag{
		{
			Key:      "mines",
			Example:  "3",
			Required: true,
		},
		{
			Key:      "positions",
			Example:  "0,1,2",
			Required: true,
		},
		{
			Key:      "seed",
			Example:  "{base64_string}",
			Required: false,
		},
	}
}

func (parser *MinesArgParser) Parse(ctx context.Context, args map[string]string) (func() games.Controller, error) {
	logger, ok := ctx.Value(contextx.CtxKeyLogger).(logsx.Logger)
	if !ok {
		logger = logsx.NopLogger{}
	}

	var mines int
	var positions []int
	var seed *string
	for key, value := range args {
		var err error
		switch key {
		case "mines":
			mines, err = strconv.Atoi(value)
			if err != nil {
				return nil, errors.New("failed to parse 'mines' args. Error:" + err.Error())
			}
		case "positions":
			buff := bytes.NewBuffer(make([]byte, 0, 50))

			appendFromBuff := func(i int) error {
				if buff.Len() == 0 {
					return fmt.Errorf("position is missing at index %d", i)
				}

				pos, err := strconv.Atoi(buff.String())
				if err != nil {
					return fmt.Errorf("failed to convert position at index %d", i)
				}

				positions = append(positions, pos)
				buff.Reset()

				return nil
			}

			posIndex := 0
			for i, char := range value {
				switch {
				case char == ' ':
					continue
				case char == ',':
					if err = appendFromBuff(posIndex); err != nil {
						return nil, err
					}
					posIndex++
				case '0' <= char && char <= '9':
					buff.WriteRune(char)
				default:
					err = fmt.Errorf("invalid character (%b) for position at index %d", char, i)
				}
			}

			if err = appendFromBuff(posIndex); err != nil {
				return nil, err
			}
		case "seed":
			seed = new(value)
		}
	}

	if mines <= 0 || len(positions) == 0 {
		return nil, errors.New("missing mines arguments")
	}

	logger.DebugContext(ctx, "Mines params", slog.Int("mine", mines), slog.Any("positions", positions))

	return func() games.Controller {
		return &minescontroller.Controller{
			Mines:     mines,
			Positions: positions,
			Logger:    logger,
			Seed:      seed,
		}
	}, nil
}
