package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const (
	betSizesKey  = "games.bet_sizes"
	maxPayoutKey = "games.max_payout_multiplier"
)

type Games struct {
	BetSizes            []int64
	MaxPayoutMultiplier int64
}

func (b *Games) load() error {
	betSizesStr := viper.GetString(betSizesKey)
	if betSizesStr == "" {
		return errors.New(betSizesKey + " is missing")
	}
	betsSplit := strings.Split(betSizesStr, ",")
	sizes := make([]int64, len(betsSplit))
	for i, betStr := range betsSplit {
		bet, err := strconv.Atoi(betStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s entry at index: %d. Value: %s", betSizesKey, i, betStr)
		}
		sizes[i] = int64(bet)
	}
	b.BetSizes = sizes

	b.MaxPayoutMultiplier = viper.GetInt64(maxPayoutKey)
	if b.MaxPayoutMultiplier == 0 {
		return errors.New(maxPayoutKey + " is zero")
	}

	return nil
}
