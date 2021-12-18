package db

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/cuducos/minha-receita/transform"
	"github.com/go-pg/pg/v10"
	"github.com/schollz/progressbar/v3"
)

const (
	tableName       = "cnpj"
	idFieldName     = "id"
	jsonFieldName   = "json"
	batchSize       = 2048
	pgCopyProcesses = 128
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

type row struct {
	ID   string
	JSON string
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(n string) (string, error) {
	sql, err := p.sqlFromTemplate("select.sql")
	if err != nil {
		return "", fmt.Errorf("error loading template: %w", err)
	}
	var r row
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

func (p *PostgreSQL) copy(batch []row) error {
	var data bytes.Buffer
	w := csv.NewWriter(&data)
	w.Write([]string{idFieldName, jsonFieldName})
	for _, r := range batch {
		w.Write([]string{r.ID, r.JSON})
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

type importTask struct {
	queue  chan string
	errors chan error
	done   chan struct{}
	bar    *progressbar.ProgressBar
}

func (p *PostgreSQL) batchCreator(t *importTask) {
	var idx int
	batch := make([]row, batchSize) // fixed size to make it quicker
	for pth := range t.queue {
		n, err := transform.CNPJForPath(pth)
		if err != nil {
			t.errors <- fmt.Errorf("error getting cnpj for path %s: %w", pth, err)
			return
		}
		b, err := ioutil.ReadFile(pth)
		if err != nil {
			t.errors <- fmt.Errorf("error reading %s: %w", pth, err)
			return
		}
		batch[idx] = row{n, strings.TrimSpace(string(b))}
		idx++

		if idx == batchSize {
			if err := p.copy(batch); err != nil {
				t.errors <- fmt.Errorf("error calling copy command: %w", err)
				return
			}
			t.bar.Add(len(batch))
			batch = make([]row, batchSize)
			idx = 0
		}
	}

	// remove zero-values from the remaining batch before calling pgCopy
	var c []row
	for _, r := range batch {
		if r.ID == "" {
			break
		}
		c = append(c, r)
	}
	if len(c) > 0 {
		if err := p.copy(c); err != nil {
			t.errors <- fmt.Errorf("error calling copy command: %w", err)
		}
		t.bar.Add(len(c))
	}
	t.done <- struct{}{}
}

// ImportData reads data from JSON directory and imports it.
func (p *PostgreSQL) ImportData(dir string) error {
	t := importTask{
		make(chan string),
		make(chan error),
		make(chan struct{}),
		progressbar.Default(-1, "Writing CNPJ data to PostgreSQL"),
	}
	for i := 0; i < pgCopyProcesses; i++ {
		go p.batchCreator(&t)
	}
	go allJSONFiles(dir, t.queue, t.errors)

	var c int
	for {
		select {
		case err := <-t.errors:
			return fmt.Errorf("error running import data: %w", err)
		case <-t.done:
			c++
			if c == pgCopyProcesses {
				return nil
			}
		}
	}
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
