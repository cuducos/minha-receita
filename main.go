package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cuducos/minha-receita/api"
	"github.com/cuducos/minha-receita/db"
)

const help = `Minha Receita.

Toolbox to manage Minha Receita, including tools to handle data, database and
to spin up the web server.

Requires a POSTGRES_URI environment variable with PostgreSQL credentials.

An optional POSTGRES_SCHEMA environment variable can be user to set a different
schema than “public” (which is the default).

See --help for more details.
`

const downloadHelper = `
Faça o download desses quatro arquivos, três da Receita Federal e um do IBGE.
Salve-os no diretório data/.

Receita Federal
===============

Primeiro faço o download dos arquivos da Receita Federal convertidos para CSV e
disponibilizados pelo Brasil.IO:
https://data.brasil.io/dataset/socios-brasil/_meta/list.html

Precisamos apenas desses três arquivos:

  * empresa.csv.gz
  * socio.csv.gz
  * cnae-secundaria.csv.gz

IBGE
====

Depois, faça o download do arquivo a descrição dos CNAE (Classificação Nacional
de Atividades Econômicas):
https://cnae.ibge.gov.br/classificacoes/download-concla.html

Precisamos apenas do arquivo:

  * CNAE_Subclasses_2_3_Estrutura_Detalhada.xlsx`

const apiHelper = `
Starts the web API.

The port used is 8000, unless an environment variable PORT points to a 
different number.

Using GODEBUG environment variable changes the HTTP server verbosity (for
example: http2debug=1 is verbose and http2debug=2 is more verbose, as in
https://golang.org/pkg/net/http/`

var dir string
var pg db.PostgreSQL

var rootCmd = &cobra.Command{
	Use:   "minha-receita <command>",
	Short: "Minha Receita toolbox.",
	Long:  help,
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Spins up the web API.",
	Long:  apiHelper,
	Run: func(_ *cobra.Command, _ []string) {
		api.Serve(&pg)
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads the required compressed CSV and Excel files.",
	Long:  "Downloads the required compressed CSV and Excel files.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(downloadHelper)
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates the required tables in PostgreSQL.",
	Long:  "Creates the required tables in PostgreSQL, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		pg.CreateTables()
	},
}

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drops the tables in PostgreSQL.",
	Long:  "Drops the tables in PostgreSQL, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		pg.DropTables()
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports data into PostgreSQL.",
	Long:  "Reads the compressed CSV and Excel files from a directory and copy their contents to the PostgreSQL tables, using the environment variable POSTGRES_URI to connect to the database.",
	Run: func(_ *cobra.Command, _ []string) {
		pg.ImportData(dir)
	},
}

func main() {

	downloadCmd.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	importCmd.Flags().StringVarP(&dir, "directory", "d", "data", "data directory")
	pg = db.NewPostgreSQL()

	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(dropCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
