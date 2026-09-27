package main

import (
	"os"

	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/config"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/engine"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := logsx.NewLogger(os.Stdout)

	engine.Start(cfg, logger)
}
