package db

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/go-pg/pg/v10"
)

const (
	tableName             = "cnpj"
	idFieldName           = "id"
	jsonFieldName         = "json"
	partnersJSONFieldName = "qsa"
	batchSize             = 2048
)

//go:embed postgres
var sql embed.FS

// PostgreSQL database interface.
type PostgreSQL struct {
	conn                  *pg.DB
	uri                   string
	schema                string
	sql                   map[string]string
	TableName             string
	IDFieldName           string
	JSONFieldName         string
	PartnersJSONFieldName string
}

func (p *PostgreSQL) loadTemplates() error {
	ls, err := filepath.Glob(filepath.Join("..", "db", "postgres", "*.sql"))
	if err != nil {
		return fmt.Errorf("error looking for templates: %w", err)
	}
	for _, n := range ls {
		t, err := template.ParseFS(sql, filepath.Join("postgres", filepath.Base(n)))
		if err != nil {
			return fmt.Errorf("error parsing %s template: %w", n, err)
		}
		var b bytes.Buffer
		if err = t.Execute(&b, p); err != nil {
			return fmt.Errorf("error rendering %s template: %w", n, err)
		}
		p.sql[filepath.Base(n)] = b.String()
	}
	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgreSQL) Close() { p.conn.Close() }

// TableFullName is the name of the schame and table in dot-notation.
func (p *PostgreSQL) TableFullName() string {
	return fmt.Sprintf("%s.%s", p.schema, p.TableName)
}

// CreateTable creates the required database table.
func (p *PostgreSQL) CreateTable() error {
	log.Output(2, fmt.Sprintf("Creating table %s…", p.TableFullName()))
	if _, err := p.conn.Exec(p.sql["create.sql"]); err != nil {
		return fmt.Errorf("error creating table with: %s\n%w", p.sql["create.sql"], err)
	}
	return nil
}

// DropTable drops the database table created by `CreateTable`.
func (p *PostgreSQL) DropTable() error {
	log.Output(2, fmt.Sprintf("Dropping table %s…", p.TableFullName()))
	if _, err := p.conn.Exec(p.sql["drop.sql"]); err != nil {
		return fmt.Errorf("error dropping table with: %s\n%w", p.sql["drop.sql"], err)
	}
	return nil
}

// CreateCompanies performs a copy to create a batch of companies in the
// database. It expects an array and each item should be another array with only
// two items: the ID and the JSON field values.
func (p *PostgreSQL) CreateCompanies(batch [][]string) error {
	var data bytes.Buffer
	w := csv.NewWriter(&data)
	w.Write([]string{idFieldName, jsonFieldName})
	for _, r := range batch {
		w.Write([]string{r[0], r[1]})
	}
	w.Flush()

	var out bytes.Buffer
	cmd := exec.Command(
		"psql",
		p.uri,
		"-c",
		fmt.Sprintf(`\copy %s FROM STDIN DELIMITER ',' CSV HEADER;`, tableName),
	)
	cmd.Stdin = &data
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while importing data to postgres %s: %w", out.String(), err)
	}
	return nil
}

// Returns the minimum and maximum CNPJ possible given a base CNPJ.
func rangeFor(base string) (int64, int64, error) {
	n, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error converting base cnpj %s to integer: %w", base, err)
	}
	mm := int64(math.Pow(10, 6))
	min := n * mm // adds 6 zeroes to complete the CNPJ's 14 digits
	return min, min + (mm - 1), nil
}

// UpdateCompanies performs a update in the JSON from the database, merging it
// with `json`.
func (p *PostgreSQL) UpdateCompanies(base, json string) error {
	min, max, err := rangeFor(base)
	if err != nil {
		return fmt.Errorf("error calculating the cnpj interval for base %s: %w", base, err)
	}
	if _, err := p.conn.Exec(p.sql["update.sql"], json, min, max); err != nil {
		return fmt.Errorf("error updating cnpj base %s: %s\n%w", base, p.sql["update.sql"], err)
	}
	return nil
}

// AddPartner appends a partner to the existing list of partners in the database.
func (p *PostgreSQL) AddPartner(base string, json string) error {
	min, max, err := rangeFor(base)
	if err != nil {
		return fmt.Errorf("error calculating the cnpj interval for base %s: %w", base, err)
	}
	json = "[" + json + "]" // postgres expects an array, not an object
	if _, err := p.conn.Exec(p.sql["add_partner.sql"], json, json, min, max); err != nil {
		return fmt.Errorf("error listing with base %s: %s\n%w", base, p.sql["add_partner.sql"], err)
	}
	return nil
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(id string) (string, error) {
	n, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("error converting cnpj %s to integer: %w", id, err)
	}
	var row struct {
		ID   int
		JSON string
	}
	if _, err := p.conn.QueryOne(&row, p.sql["get.sql"], n); err != nil {
		return "", fmt.Errorf("error getting CNPJ %s with: %s\n%w", cnpj.Mask(id), p.sql["get.sql"], err)
	}
	return row.JSON, nil
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(uri, schema string) (PostgreSQL, error) {
	opt, err := pg.ParseURL(uri)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("unable to parse postgres uri %s: %w", uri, err)
	}
	p := PostgreSQL{
		pg.Connect(opt),
		uri,
		schema,
		make(map[string]string),
		tableName,
		idFieldName,
		jsonFieldName,
		partnersJSONFieldName,
	}
	if err = p.loadTemplates(); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not load the sql templates: %w", err)
	}
	if err := p.conn.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
