package main

import (
	"fmt"
	"os"

	"github.com/cuducos/minha-receita/cmd"
)

func main() {
	if err := cmd.CLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
