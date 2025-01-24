package main

import (
	"fmt"
	"os"

	"github.com/mbnunes/minha-receita/cmd"
)

func main() {
	if err := cmd.CLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
