package db

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/go-pg/pg/v10"
)

const (
	tableName         = "cnpj"
	idFieldName       = "id"
	baseCNPJFieldName = "base"
	jsonFieldName     = "json"
	batchSize         = 2048
)

//go:embed postgres
var sql embed.FS

// PostgreSQL database interface.
type PostgreSQL struct {
	conn              *pg.DB
	uri               string
	schema            string
	TableName         string
	IDFieldName       string
	BaseCNPJFieldName string
	JSONFieldName     string
}

// Close closes the PostgreSQL connection
func (p *PostgreSQL) Close() { p.conn.Close() }

// TableFullName is the name of the schame and table in dot-notation.
func (p *PostgreSQL) TableFullName() string {
	return fmt.Sprintf("%s.%s", p.schema, p.TableName)
}

func (p *PostgreSQL) sqlFromTemplate(n string) (string, error) {
	t, err := template.ParseFS(sql, filepath.Join("postgres", n))
	if err != nil {
		return "", fmt.Errorf("error parsing %s template: %w", n, err)
	}
	var b bytes.Buffer
	if err = t.Execute(&b, p); err != nil {
		return "", fmt.Errorf("error rendering %s template: %w", n, err)
	}
	return b.String(), nil
}

// CreateTable creates the required database table.
func (p *PostgreSQL) CreateTable() error {
	sql, err := p.sqlFromTemplate("create.sql")
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}
	log.Output(2, fmt.Sprintf("Creating table %s…", p.TableFullName()))
	if _, err := p.conn.Exec(sql); err != nil {
		return fmt.Errorf("error creating table with: %s\n%w", sql, err)
	}
	return nil
}

// DropTable drops the database table created by `CreateTable`.
func (p *PostgreSQL) DropTable() error {
	sql, err := p.sqlFromTemplate("drop.sql")
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}
	log.Output(2, fmt.Sprintf("Dropping table %s…", p.TableFullName()))
	if _, err := p.conn.Exec(sql); err != nil {
		return fmt.Errorf("error dropping table with: %s\n%w", sql, err)
	}
	return nil
}

// CreateCompanies performs a copy to create a batch of companies in the
// database. It expects an array and each item should be another array with only
// two items: the ID and the JSON field values.
func (p *PostgreSQL) CreateCompanies(batch [][]string) error {
	var data bytes.Buffer
	w := csv.NewWriter(&data)
	w.Write([]string{idFieldName, baseCNPJFieldName, jsonFieldName})
	for _, r := range batch {
		w.Write([]string{r[0], r[0][0:8], r[1]})
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

// UpdateCompanies performs a update in the JSON from the database, merging it
// with `json`.
func (p *PostgreSQL) UpdateCompanies(base, json string) error {
	sql, err := p.sqlFromTemplate("update.sql")
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}
	n, err := strconv.ParseInt(base, 10, 0)
	if err != nil {
		return fmt.Errorf("error converting base cnpj %s to integer: %w", base, err)
	}
	if _, err := p.conn.Exec(sql, json, n); err != nil {
		return fmt.Errorf("error updating cnpj base %s: %s\n%w", base, sql, err)
	}
	return nil
}

// AddPartner appends a partner to the existing list of partners in the database.
func (p *PostgreSQL) AddPartner(base string, json string) error {
	sql, err := p.sqlFromTemplate("add_partner.sql")
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}
	n, err := strconv.ParseInt(base, 10, 0)
	if err != nil {
		return fmt.Errorf("error converting base cnpj %s to integer: %w", base, err)
	}
	json = "[" + json + "]" // postgres expects an array, not an object
	if _, err := p.conn.Exec(sql, json, json, n); err != nil {
		return fmt.Errorf("error listing with base %s: %s\n%w", base, sql, err)
	}
	return nil
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(id string) (string, error) {
	sql, err := p.sqlFromTemplate("get.sql")
	if err != nil {
		return "", fmt.Errorf("error loading template: %w", err)
	}
	n, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("error converting cnpj %s to integer: %w", id, err)
	}
	var row struct {
		ID   int
		JSON string
	}
	if _, err := p.conn.QueryOne(&row, sql, n); err != nil {
		return "", fmt.Errorf("error getting CNPJ %s with: %s\n%w", cnpj.Mask(id), sql, err)
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
		tableName,
		idFieldName,
		baseCNPJFieldName,
		jsonFieldName,
	}
	if err := p.conn.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
