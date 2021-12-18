package db

import (
	"bytes"
	"compress/gzip"
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/csv"
	"github.com/go-pg/pg/v10"
)

const tableName = "cnpj"

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

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(n string) (string, error) {
	sql, err := p.sqlFromTemplate("select.sql")
	if err != nil {
		return "", fmt.Errorf("error loading template: %w", err)
	}
	var r struct {
		ID   string
		JSON string
	}
	if _, err := p.conn.QueryOne(&r, sql, n); err != nil {
		return "", fmt.Errorf("error getting CNPJ %s with: %s\n%w", cnpj.Mask(n), sql, err)
	}
	return r.JSON, nil
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
	log.Output(2, "Done!")
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
	log.Output(2, "Done!")
	return nil
}

// ImportData reads data from JSON directory and imports it.
func (p *PostgreSQL) ImportData(dir string) error {
	src := filepath.Join(dir, csv.Path)
	log.Output(2, fmt.Sprintf("Importing data from %s to %s…", src, p.TableFullName()))

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening csv %s: %w", src, err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("error reading gzip %s: %w", src, err)
	}
	defer r.Close()

	var out bytes.Buffer
	cmd := exec.Command(
		"psql",
		p.uri,
		"-c",
		fmt.Sprintf(`\copy %s FROM STDIN DELIMITER ',' CSV HEADER;`, p.TableFullName()),
	)
	cmd.Stdin = r
	cmd.Stderr = &out
	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("error while importing %s to %s: %s\n%w", src, p.TableFullName(), out.String(), err)
	}

	log.Output(2, fmt.Sprintf("Done! Imported data from %s to %s.", src, p.TableFullName()))
	return nil
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(u, s string) (PostgreSQL, error) {
	opt, err := pg.ParseURL(u)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("unable to parse postgres uri %s: %w", u, err)
	}
	p := PostgreSQL{pg.Connect(opt), u, s, tableName, csv.IDFieldName, csv.JSONFieldName}
	if err := p.conn.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
