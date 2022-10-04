package db

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/cuducos/go-cnpj"
	"github.com/go-pg/pg/v10"
)

const (
	companyTableName      = "cnpj"
	metaTableName         = "meta"
	idFieldName           = "id"
	jsonFieldName         = "json"
	keyFieldName          = "key"
	valueFieldName        = "value"
	partnersJSONFieldName = "qsa"
)

//go:embed postgres
var sql embed.FS

// PostgreSQL database interface.
type PostgreSQL struct {
	conn                  *pg.DB
	uri                   string
	schema                string
	sql                   map[string]string
	CompanyTableName      string
	MetaTableName         string
	IDFieldName           string
	JSONFieldName         string
	KeyFieldName          string
	ValueFieldName        string
	PartnersJSONFieldName string
}

func (p *PostgreSQL) loadTemplates() error {
	ls, err := sql.ReadDir("postgres")
	if err != nil {
		return fmt.Errorf("error looking for templates: %w", err)
	}
	for _, f := range ls {
		t, err := template.ParseFS(sql, filepath.Join("postgres", f.Name()))
		if err != nil {
			return fmt.Errorf("error parsing %s template: %w", f, err)
		}
		var b bytes.Buffer
		if err = t.Execute(&b, p); err != nil {
			return fmt.Errorf("error rendering %s template: %w", f, err)
		}
		p.sql[f.Name()] = b.String()
	}
	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgreSQL) Close() { p.conn.Close() }

// CompanyTableFullName is the name of the schame and table in dot-notation.
func (p *PostgreSQL) CompanyTableFullName() string {
	return fmt.Sprintf("%s.%s", p.schema, p.CompanyTableName)
}

// MetaTableFullName is the name of the schame and table in dot-notation.
func (p *PostgreSQL) MetaTableFullName() string {
	return fmt.Sprintf("%s.%s", p.schema, p.MetaTableName)
}

// CreateTable creates the required database table.
func (p *PostgreSQL) CreateTable() error {
	log.Output(2, fmt.Sprintf("Creating table %s…", p.CompanyTableFullName()))
	if _, err := p.conn.Exec(p.sql["create.sql"]); err != nil {
		return fmt.Errorf("error creating table with: %s\n%w", p.sql["create.sql"], err)
	}
	return nil
}

// DropTable drops the database table created by `CreateTable`.
func (p *PostgreSQL) DropTable() error {
	log.Output(2, fmt.Sprintf("Dropping table %s…", p.CompanyTableFullName()))
	if _, err := p.conn.Exec(p.sql["drop.sql"]); err != nil {
		return fmt.Errorf("error dropping table with: %s\n%w", p.sql["drop.sql"], err)
	}
	return nil
}

// AssertPostgresCLIExists searches for the PostgreSQL executable (psql) in the
// environment's PATH. It will return an error if no executable is found.
func AssertPostgresCLIExists() error {
	_, err := exec.LookPath("psql")
	if err != nil {
		return errors.New("postgres client (psql) not installed or not in PATH")
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
		fmt.Sprintf(`\copy %s FROM STDIN DELIMITER ',' CSV HEADER;`, p.CompanyTableName),
	)
	cmd.Stdin = &data
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while importing data to postgres %s: %w", out.String(), err)
	}
	return nil
}

// CreateIndex runs after all the data is creates. It drops duplicates and
// create a primary key on the ID field.
func (p *PostgreSQL) CreateIndex() error {
	log.Output(2, "Creating indexess…")
	if _, err := p.conn.Exec(p.sql["create_index.sql"]); err != nil {
		return fmt.Errorf("error creating index with: %s\n%w", p.sql["create_index.sql"], err)
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
// with `json`. It expects an array of two-items array containing a base CNPJ
// and the new JSON data.
func (p *PostgreSQL) UpdateCompanies(data [][]string) error {
	args := make([]interface{}, len(data)*3)
	for i, v := range data {
		min, max, err := rangeFor(v[0])
		if err != nil {
			return fmt.Errorf("error calculating the cnpj interval for base %s: %w", v[0], err)
		}
		args[i*3] = v[1]
		args[(i*3)+1] = min
		args[(i*3)+2] = max
	}
	query := strings.Repeat(p.sql["update.sql"], len(data))
	if _, err := p.conn.Exec(query, args...); err != nil {
		return fmt.Errorf("error updating companies: %w", err)
	}
	return nil
}

// AddPartners appends an array of partners to the existing list of partners in
// the database. It expects an array of two-items array containing a base CNPJ
// and the new JSON data.
func (p *PostgreSQL) AddPartners(data [][]string) error {
	args := make([]interface{}, len(data)*4)
	for i, v := range data {
		min, max, err := rangeFor(v[0])
		if err != nil {
			return fmt.Errorf("error calculating the cnpj interval for base %s: %w", v[0], err)
		}
		args[(i * 4)] = v[1]
		args[(i*4)+1] = v[1]
		args[(i*4)+2] = min
		args[(i*4)+3] = max
	}
	query := strings.Repeat(p.sql["add_partner.sql"], len(data))
	if _, err := p.conn.Exec(query, args...); err != nil {
		return fmt.Errorf("error adding partners: %w", err)
	}
	return nil
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(id string) (string, error) {
	n, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("error converting cnpj %s to integer: %w", id, err)
	}
	var row struct{ JSON string }
	if _, err := p.conn.QueryOne(&row, p.sql["get.sql"], n); err != nil {
		return "", fmt.Errorf("error getting CNPJ %s with: %s\n%w", cnpj.Mask(id), p.sql["get.sql"], err)
	}
	return row.JSON, nil
}

// MetaSave saves a key/value pair in the metadata table.
func (p *PostgreSQL) MetaSave(k, v string) error {
	if len(k) > 16 {
		return fmt.Errorf("metatable can only take keys that are at maximum 16 chars long")
	}
	if _, err := p.conn.Exec(p.sql["meta_save.sql"], k, v, v); err != nil {
		return fmt.Errorf("error saving %s to metadata: %w", k, err)
	}
	return nil
}

// MetaRead reads a key/value pair from the metadata table.
func (p *PostgreSQL) MetaRead(k string) string {
	var row struct{ Value string }
	if _, err := p.conn.QueryOne(&row, p.sql["meta_read.sql"], k); err != nil {
		return ""
	}
	return row.Value
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(uri, schema string) (PostgreSQL, error) {
	opt, err := pg.ParseURL(uri)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("unable to parse postgres uri %s: %w", uri, err)
	}
	p := PostgreSQL{
		conn:                  pg.Connect(opt),
		uri:                   uri,
		schema:                schema,
		sql:                   make(map[string]string),
		CompanyTableName:      companyTableName,
		MetaTableName:         metaTableName,
		IDFieldName:           idFieldName,
		JSONFieldName:         jsonFieldName,
		KeyFieldName:          keyFieldName,
		ValueFieldName:        valueFieldName,
		PartnersJSONFieldName: partnersJSONFieldName,
	}
	if err = p.loadTemplates(); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not load the sql templates: %w", err)
	}
	if err := p.conn.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
