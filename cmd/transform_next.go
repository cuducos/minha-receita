package cmd

import (
	"github.com/cuducos/minha-receita/transformnext"
	"github.com/spf13/cobra"
)

var transformNextCmd = &cobra.Command{
	Use:   "transform-next",
	Short: "New 2-step ETL process (experimental)",
	Long: `New 2-step ETL process that simplifies the original 4-step transform:

Step 1: Load all auxiliary data (Cnaes, Empresas, Socios, etc.) into Badger key-value storage
Step 2: Process Estabelecimentos with enrichment from Badger and batch database writes

This addresses the complexity issues of the original transform package:
- Centralizes all key-value data in Badger (no more in-memory lookups)
- Single enrichment step when processing Estabelecimentos
- Reduced memory usage with optimized serialization
- Eliminates unnecessary data conversions

Usage:
  DEBUG=1 minha-receita transform-next /path/to/data

Requires DATABASE_URL environment variable to be set for Step 2.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return transformnext.Transform(dir)
	},
}

var cleanupTempCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean-up temporary ETL files from transform-next",
	Long: `Removes temporary directories created by transform-next ETL process.
These directories are typically in /tmp and follow the pattern "minha-receita-*".`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return transformnext.Cleanup()
	},
}
