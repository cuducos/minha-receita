package main

import (
	"log/slog"
	"os"

	"github.com/cuducos/minha-receita/cmd"
)

func main() {
	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err := cmd.CLI().Execute(); err != nil {
		slog.Error("Exiting minha-receita", "error", err)
		os.Exit(1)
	}
}
