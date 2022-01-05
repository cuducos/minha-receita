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
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/go-pg/pg/v10"
)

const (
	tableName     = "cnpj"
	idFieldName   = "id"
	jsonFieldName = "json"
	batchSize     = 2048
)

//go:embed postgres
var sql embed.FS

// PostgreSQL database interface.
type PostgreSQL struct {
	conn          *pg.DB
	uri           string
	schema        string
	TableName     string
	IDFieldName   string
	JSONFieldName string
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

// UpdateCompany performs a update in the database.
func (p *PostgreSQL) UpdateCompany(id, json string) error {
	sql, err := p.sqlFromTemplate("update.sql")
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}
	if _, err := p.conn.Exec(sql, json, id); err != nil {
		return fmt.Errorf("error updating record %s: %s\n%w", cnpj.Mask(id), sql, err)
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
		w.Write(r)
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

type row struct {
	ID   string
	JSON string
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(n string) (string, error) {
	sql, err := p.sqlFromTemplate("get.sql")
	if err != nil {
		return "", fmt.Errorf("error loading template: %w", err)
	}
	var r row
	if _, err := p.conn.QueryOne(&r, sql, n); err != nil {
		return "", fmt.Errorf("error getting CNPJ %s with: %s\n%w", cnpj.Mask(n), sql, err)
	}
	return r.JSON, nil
}

// ListCompanies returns the JSON for all companies with a CNPJ starting with a `base`.
func (p *PostgreSQL) ListCompanies(base string) ([]string, error) {
	sql, err := p.sqlFromTemplate("list.sql")
	if err != nil {
		return []string{}, fmt.Errorf("error loading template: %w", err)
	}
	var j []string
	if _, err := p.conn.Query(&j, sql, base+"%"); err != nil {
		return []string{}, fmt.Errorf("error listing with base %s: %s\n%w", base, sql, err)
	}
	return j, nil
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(u, s string) (PostgreSQL, error) {
	opt, err := pg.ParseURL(u)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("unable to parse postgres uri %s: %w", u, err)
	}
	p := PostgreSQL{pg.Connect(opt), u, s, tableName, idFieldName, jsonFieldName}
	if err := p.conn.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
