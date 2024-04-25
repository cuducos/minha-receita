package db

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newrelic/go-agent/v3/integrations/nrpgx5"
	"github.com/newrelic/go-agent/v3/newrelic"
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
	pool                  *pgxpool.Pool
	newRelic              *newrelic.Application
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
		t, err := template.ParseFS(sql, "postgres/"+f.Name())
		if err != nil {
			return fmt.Errorf("error parsing %s template: %w", f, err)
		}
		var b bytes.Buffer
		if err = t.Execute(&b, p); err != nil {
			return fmt.Errorf("error rendering %s template: %w", f, err)
		}
		p.sql[strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))] = b.String()
	}
	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgreSQL) Close() { p.pool.Close() }

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
	log.Output(1, fmt.Sprintf("Creating table %s…", p.CompanyTableFullName()))
	if _, err := p.pool.Exec(context.Background(), p.sql["create"]); err != nil {
		return fmt.Errorf("error creating table with: %s\n%w", p.sql["create"], err)
	}
	return nil
}

// DropTable drops the database table created by `CreateTable`.
func (p *PostgreSQL) DropTable() error {
	log.Output(1, fmt.Sprintf("Dropping table %s…", p.CompanyTableFullName()))
	if _, err := p.pool.Exec(context.Background(), p.sql["drop"]); err != nil {
		return fmt.Errorf("error dropping table with: %s\n%w", p.sql["drop"], err)
	}
	return nil
}

// CreateCompanies performs a copy to create a batch of companies in the
// database. It expects an array and each item should be another array with only
// two items: the ID and the JSON field values.
func (p *PostgreSQL) CreateCompanies(batch [][]any) error {
	_, err := p.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{p.CompanyTableName},
		[]string{idFieldName, jsonFieldName},
		pgx.CopyFromRows(batch),
	)
	if err != nil {
		return fmt.Errorf("error while importing data to postgres: %w", err)
	}
	return nil
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(id string) (string, error) {
	n, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("error converting cnpj %s to integer: %w", id, err)
	}

	ctx := context.Background()
	if p.newRelic != nil {
		txn := p.newRelic.StartTransaction("GetCompany")
		ctx = newrelic.NewContext(ctx, txn)
		defer txn.End()
	}

	rows, err := p.pool.Query(ctx, p.sql["get"], n)
	if err != nil {
		return "", fmt.Errorf("error looking for cnpj %d: %w", n, err)
	}
	j, err := pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return "", fmt.Errorf("error reading cnpj %d: %w", n, err)
	}
	return j, nil
}

// PreLoad runs before starting to load data into the database. Currently it
// disables autovacuum on PostgreSQL.
func (p *PostgreSQL) PreLoad() error {
	if _, err := p.pool.Exec(context.Background(), p.sql["pre_load"]); err != nil {
		return fmt.Errorf("error during pre load: %s\n%w", p.sql["pre_load"], err)
	}
	return nil
}

// PostLoad runs after loading data into the database. Currently it re-enables
// autovacuum on PostgreSQL.
func (p *PostgreSQL) PostLoad() error {
	if _, err := p.pool.Exec(context.Background(), p.sql["post_load"]); err != nil {
		return fmt.Errorf("error during post load: %s\n%w", p.sql["autovacuum"], err)
	}
	return nil
}

// MetaSave saves a key/value pair in the metadata table.
func (p *PostgreSQL) MetaSave(k, v string) error {
	if len(k) > 16 {
		return fmt.Errorf("metatable can only take keys that are at maximum 16 chars long")
	}
	if _, err := p.pool.Exec(context.Background(), p.sql["meta_save"], k, v); err != nil {
		return fmt.Errorf("error saving %s to metadata: %w", k, err)
	}
	return nil
}

// MetaRead reads a key/value pair from the metadata table.
func (p *PostgreSQL) MetaRead(k string) (string, error) {
	rows, err := p.pool.Query(context.Background(), p.sql["meta_read"], k)
	if err != nil {
		return "", fmt.Errorf("error looking for metadata key %s: %w", k, err)
	}
	v, err := pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return "", fmt.Errorf("error reading for metadata key %s: %w", k, err)
	}
	return v, nil
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(uri, schema string, nr *newrelic.Application) (PostgreSQL, error) {
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("could not create database config: %w", err)
	}
	if nr != nil {
		cfg.ConnConfig.Tracer = nrpgx5.NewTracer()
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to the database: %w", err)
	}
	p := PostgreSQL{
		pool:                  conn,
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
	if err := p.pool.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
