package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cuducos/minha-receita/cmd"
)

func main() {
	if os.Getenv("DEBUG") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err := cmd.CLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
