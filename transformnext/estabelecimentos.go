package transformnext

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cuducos/minha-receita/db"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

type enrichedCompany struct {
	cnpj             string
	companyData      []string
	venueData        []string
	partnersData     [][]string
	taxesData        [][]string
	cnaesSecundarios [][]string
}

func enrichCompanyData(cnpj string, kv *kv) (*enrichedCompany, error) {
	company := &enrichedCompany{cnpj: cnpj}

	// Get company data from Empresas
	if companyData, err := kv.get([]byte(fmt.Sprintf("%s::emp", cnpj))); err == nil && companyData != nil {
		company.companyData = companyData
	}

	// Get venue data from Estabelecimentos (this will be the main row)
	// Note: In the actual implementation, this comes from the current CSV row being processed

	// Get partners data
	if partnersData, err := kv.getPrefix([]byte(fmt.Sprintf("%s::soc", cnpj))); err == nil {
		company.partnersData = partnersData
	}

	// Get taxes data (all tax regimes)
	taxSources := []string{"ari", "lpr", "lre", "sim"}
	for _, tax := range taxSources {
		if taxData, err := kv.getPrefix([]byte(fmt.Sprintf("%s::%s", cnpj, tax))); err == nil {
			company.taxesData = append(company.taxesData, taxData...)
		}
	}

	// Get secondary CNAEs
	if cnaesData, err := kv.getPrefix([]byte(fmt.Sprintf("%s::cna", cnpj))); err == nil {
		company.cnaesSecundarios = cnaesData
	}

	return company, nil
}

func processEstabelecimentos(dir string, kv *kv) error {
	slog.Info("Step 2: Processing Estabelecimentos with enrichment")

	// Find Estabelecimentos files
	estabelecimentosFiles, err := findEstabelecimentosFiles(dir)
	if err != nil {
		return fmt.Errorf("could not find Estabelecimentos files: %w", err)
	}

	if len(estabelecimentosFiles) == 0 {
		return fmt.Errorf("no Estabelecimentos files found in %s", dir)
	}

	// Create database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}
	dbConn, err := db.NewPostgreSQL(dbURL, "public")
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}
	defer dbConn.Close()

	// Prepare database
	if err := dbConn.PreLoad(); err != nil {
		return fmt.Errorf("could not prepare database: %w", err)
	}

	bar := progressbar.NewOptions(
		len(estabelecimentosFiles),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription("[Step 2 of 2] Processing Estabelecimentos"),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionShowCount(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var g errgroup.Group
	for _, file := range estabelecimentosFiles {
		file := file
		g.Go(func() error {
			return processEstabelecimentosFile(ctx, file, kv, &dbConn, bar)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Post-load operations
	if err := dbConn.PostLoad(); err != nil {
		return fmt.Errorf("could not complete post-load operations: %w", err)
	}

	slog.Info("Step 2 completed successfully")
	return nil
}

func findEstabelecimentosFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "Estabelecimentos") && strings.HasSuffix(entry.Name(), ".zip") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

func processEstabelecimentosFile(ctx context.Context, filePath string, kv *kv, dbConn *db.PostgreSQL, bar *progressbar.ProgressBar) error {
	defer func() {
		if err := bar.Add(1); err != nil {
			slog.Warn("could not update progress bar", "error", err)
		}
	}()

	slog.Debug("Processing Estabelecimentos file", "file", filePath)

	// Open zip file
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("could not open zip file %s: %w", filePath, err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			slog.Warn("could not close archive", "file", filePath, "error", err)
		}
	}()

	var wg sync.WaitGroup
	batchChan := make(chan []string, 100) // Buffered channel for batches

	// Start batch processor
	var processErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		processErr = processBatch(batchChan, kv, dbConn)
	}()

	// Process each CSV file in the zip
	for _, zipFile := range archive.File {
		if !strings.HasSuffix(zipFile.Name, ".csv") {
			continue
		}

		if err := processCSVFile(ctx, zipFile, kv, batchChan); err != nil {
			close(batchChan)
			return fmt.Errorf("error processing CSV file %s: %w", zipFile.Name, err)
		}
	}

	close(batchChan)
	wg.Wait()

	return processErr
}

func processCSVFile(ctx context.Context, zipFile *zip.File, kv *kv, batchChan chan<- []string) error {
	fileReader, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", zipFile.Name, err)
	}
	defer func() {
		if err := fileReader.Close(); err != nil {
			slog.Warn("could not close file reader", "file", zipFile.Name, "error", err)
		}
	}()

	reader := csv.NewReader(charmap.ISO8859_15.NewDecoder().Reader(fileReader))
	reader.Comma = ';'

	// Skip header if present
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("could not read header: %w", err)
	}

	batch := make([]string, 0, 8192) // BatchSize from original

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			row, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					goto doneReading
				}
				return fmt.Errorf("error reading CSV row: %w", err)
			}

			// Clean up row data
			for i, field := range row {
				row[i] = cleanupColumn(field)
			}

			if len(row) == 0 || row[0] == "" {
				continue
			}

			// Enrich the data
			enriched, err := enrichCompanyData(row[0], kv)
			if err != nil {
				slog.Warn("could not enrich company data", "cnpj", row[0], "error", err)
				continue
			}

			// Use the venue data from the current row
			enriched.venueData = row[1:]

			// Convert to JSON
			jsonData, err := convertToJSON(enriched, kv)
			if err != nil {
				slog.Warn("could not convert to JSON", "cnpj", row[0], "error", err)
				continue
			}

			batch = append(batch, jsonData)

			// Send batch when full
			if len(batch) >= 8192 {
				batchCopy := make([]string, len(batch))
				copy(batchCopy, batch)
				batchChan <- batchCopy
				batch = batch[:0] // Reset batch
			}
		}
	}

doneReading:

	// Send remaining batch
	if len(batch) > 0 {
		batchChan <- batch
	}

	return nil
}

func processBatch(batchChan <-chan []string, kv *kv, dbConn *db.PostgreSQL) error {
	for batch := range batchChan {
		// Convert []string to [][]string for CreateCompanies
		companies := make([][]string, len(batch))
		for i, jsonStr := range batch {
			companies[i] = []string{jsonStr}
		}
		if err := dbConn.CreateCompanies(companies); err != nil {
			return fmt.Errorf("could not create companies in database: %w", err)
		}
	}
	return nil
}

func convertToJSON(company *enrichedCompany, kv *kv) (string, error) {
	// This is a simplified JSON conversion
	// In the real implementation, this would build the complete JSON structure
	// with all the enriched data following the expected format

	// For now, return a placeholder that shows the structure
	jsonStr := fmt.Sprintf(`{
		"cnpj": "%s",
		"company_data": %v,
		"venue_data": %v,
		"partners": %v,
		"taxes": %v,
		"secondary_cnaes": %v
	}`, company.cnpj, company.companyData, company.venueData, company.partnersData, company.taxesData, company.cnaesSecundarios)

	return jsonStr, nil
}
