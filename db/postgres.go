package db

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/cuducos/minha-receita/transform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	companyTableName = "cnpj"
	metaTableName    = "meta"
	cursorFieldName  = "cursor"
	idFieldName      = "id"
	jsonFieldName    = "json"
	keyFieldName     = "key"
	valueFieldName   = "value"
)

//go:embed postgres
var sql embed.FS

type sqlTemplate struct {
	path         fs.DirEntry
	embeddedPath string
	key          string
}

func (s *sqlTemplate) render(p *PostgreSQL) (string, error) {
	t, err := template.ParseFS(sql, s.embeddedPath)
	if err != nil {
		return "", fmt.Errorf("error parsing %s template: %w", s.path, err)
	}
	var b bytes.Buffer
	if err = t.Execute(&b, p); err != nil {
		return "", fmt.Errorf("error rendering %s template: %w", s.path, err)
	}
	return b.String(), nil

}

func newSQLTemplate(f fs.DirEntry) sqlTemplate {
	return sqlTemplate{
		path:         f,
		embeddedPath: "postgres/" + f.Name(),
		key:          strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())),
	}
}

type ExtraIndex struct {
	IsRoot bool
	Name   string
	Value  string
}

func (e *ExtraIndex) NestedPath() string {
	if e.IsRoot {
		slog.Error("cannot not parse nested path for index at the root of the json", "index", e.Value)
		return ""
	}
	p := strings.SplitN(e.Value, ".", 2)
	if len(p) != 2 {
		slog.Error("could not parse nested path", "index", e.Value)
		return ""
	}
	return fmt.Sprintf("$.%s[*].%s", p[0], p[1])
}

// PostgreSQL database interface.
type PostgreSQL struct {
	pool             *pgxpool.Pool
	uri              string
	schema           string
	getCompanyQuery  string
	metaReadQuery    string
	CompanyTableName string
	MetaTableName    string
	CursorFieldName  string
	IDFieldName      string
	JSONFieldName    string
	KeyFieldName     string
	ValueFieldName   string
	ExtraIndexes     []ExtraIndex
}

func (p *PostgreSQL) renderTemplate(key string) (string, error) {
	ls, err := sql.ReadDir("postgres")
	if err != nil {
		return "", fmt.Errorf("error looking for templates: %w", err)
	}
	for _, f := range ls {
		s := newSQLTemplate(f)
		if s.key != key {
			continue
		}
		return s.render(p)
	}
	return "", fmt.Errorf("template %s not found", key)
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

// Create creates the required database table.
func (p *PostgreSQL) Create() error {
	slog.Info("Creating", "table", p.CompanyTableFullName())
	s, err := p.renderTemplate("create")
	if err != nil {
		return fmt.Errorf("error rendering create template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s); err != nil {
		return fmt.Errorf("error creating table with: %s\n%w", s, err)
	}
	return nil
}

// Drop drops the database table created by `Create`.
func (p *PostgreSQL) Drop() error {
	slog.Info("Dropping", "table", p.CompanyTableFullName())
	s, err := p.renderTemplate("drop")
	if err != nil {
		return fmt.Errorf("error rendering drop template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s); err != nil {
		return fmt.Errorf("error dropping table with: %s\n%w", s, err)
	}
	return nil
}

// CreateCompanies performs a copy to create a batch of companies in the
// database. It expects an array and each item should be another array with only
// two items: the ID and the JSON field values.
func (p *PostgreSQL) CreateCompanies(batch [][]string) error {
	b := make([][]any, len(batch))
	for i, r := range batch {
		b[i] = []any{r[0], r[1]}
	}
	_, err := p.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{p.CompanyTableName},
		[]string{idFieldName, jsonFieldName},
		pgx.CopyFromRows(b),
	)
	if err != nil {
		return fmt.Errorf("error while importing data to postgres: %w", err)
	}
	return nil
}

// GetCompany returns the JSON of a company based on a CNPJ number.
func (p *PostgreSQL) GetCompany(id string) (string, error) {
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, p.getCompanyQuery, id)
	if err != nil {
		return "", fmt.Errorf("error looking for cnpj %s: %w", id, err)
	}
	j, err := pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return "", fmt.Errorf("error reading cnpj %s: %w", id, err)
	}
	return j, nil
}

func (p *PostgreSQL) searchQuery(q *Query) squirrel.SelectBuilder {
	b := squirrel.
		Select(p.CursorFieldName, p.JSONFieldName).
		From(p.CompanyTableFullName()).
		OrderBy(p.CursorFieldName).
		Limit(uint64(q.Limit))
	if q.Cursor != nil {
		c, err := q.CursorAsInt()
		if err == nil {
			b = b.Where(squirrel.Gt{p.CursorFieldName: c})
		}
	}
	if len(q.UF) > 0 {
		cs := make([]squirrel.Sqlizer, len(q.UF))
		for i, v := range q.UF {
			cs[i] = squirrel.Expr(fmt.Sprintf(`json -> 'uf' = '"%s"'::jsonb`, v))
		}
		b = b.Where(squirrel.Or(cs))
	}
	if len(q.Municipio) > 0 {
		cs := make([]squirrel.Sqlizer, len(q.Municipio))
		for i, v := range q.Municipio {
			ms := make([]squirrel.Sqlizer, 2)
			ms[0] = squirrel.Expr(fmt.Sprintf("json -> 'codigo_municipio' = '%d'::jsonb", v))
			ms[1] = squirrel.Expr(fmt.Sprintf("json -> 'codigo_municipio_ibge' = '%d'::jsonb", v))
			cs[i] = squirrel.Or(ms)
		}
		b = b.Where(squirrel.Or(cs))
	}
	if len(q.NaturezaJuridica) > 0 {
		cs := make([]squirrel.Sqlizer, len(q.NaturezaJuridica))
		for i, v := range q.NaturezaJuridica {
			cs[i] = squirrel.Expr(fmt.Sprintf("json -> 'codigo_natureza_juridica' = '%d'::jsonb", v))
		}
		b = b.Where(squirrel.Or(cs))
	}
	if len(q.CNAEFiscal) > 0 {
		cs := make([]squirrel.Sqlizer, len(q.CNAEFiscal))
		for i, v := range q.CNAEFiscal {
			cs[i] = squirrel.Expr(fmt.Sprintf("json -> 'cnae_fiscal' = '%d'::jsonb", v))
		}
		b = b.Where(squirrel.Or(cs))
	}
	if len(q.CNAE) > 0 {
		cs := make([]squirrel.Sqlizer, len(q.CNAE)+1)
		s := make([]string, len(q.CNAE))
		for i, v := range q.CNAE {
			s[i] = fmt.Sprintf("%d", v)
			cs[i] = squirrel.Expr(fmt.Sprintf("json -> 'cnae_fiscal' = '%d'::jsonb", v))
		}
		cs[len(q.CNAE)] = squirrel.Expr(fmt.Sprintf(
			"jsonb_path_query_array(json, '$.cnaes_secundarios[*].codigo') @> '[%s]'",
			strings.Join(s, ","),
		))
		b.Where(squirrel.Or(cs))
	}
	if len(q.CNPF) > 0 {
		s := make([]string, len(q.CNPF))
		for i, v := range q.CNPF {
			s[i] = fmt.Sprintf(`"%s"`, v)
		}
		b = b.Where(squirrel.Expr(fmt.Sprintf(
			"jsonb_path_query_array(json, '$.qsa[*].cnpj_cpf_do_socio') @> '[%s]'",
			strings.Join(s, ","),
		)))
	}
	return b
}

type postgresRecord struct {
	Cursor  int
	Company string
}

// Search returns paginated results with JSON for companies bases on a search
// query
func (p *PostgreSQL) Search(ctx context.Context, q *Query) (string, error) {
	s, a, err := p.searchQuery(q).ToSql()
	if err != nil {
		return "", fmt.Errorf("error building the query: %w", err)
	}
	slog.Debug("search", "query", s, "args", a)
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error starting a database transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		return "", fmt.Errorf("error disabling sequential scans: %w", err)
	}
	rows, err := p.pool.Query(ctx, s, a...)
	if err != nil {
		return "", fmt.Errorf("error searching for %#v: %w", q, err)
	}
	rs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[postgresRecord])
	if err != nil {
		return "", fmt.Errorf("error reading search result for %#v: %w", q, err)
	}
	if err := tx.Commit(ctx); err != nil {
		slog.Error("error committing the read-only search transaction", "error", err)
	}
	var cs []string
	for _, r := range rs {
		cs = append(cs, r.Company)
	}
	var cur string
	if len(rs) == int(q.Limit) {
		cur = fmt.Sprintf("%d", rs[len(rs)-1].Cursor)
	}
	return newPage(cs, cur), nil

}

// PreLoad runs before starting to load data into the database. Currently it
// disables autovacuum on PostgreSQL.
func (p *PostgreSQL) PreLoad() error {
	s, err := p.renderTemplate("pre_load")
	if err != nil {
		return fmt.Errorf("error rendering pre-load template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s); err != nil {
		return fmt.Errorf("error during pre load: %s\n%w", s, err)
	}
	return nil
}

// PostLoad runs after loading data into the database. Currently it re-enables
// autovacuum on PostgreSQL.
func (p *PostgreSQL) PostLoad() error {
	s, err := p.renderTemplate("post_load")
	if err != nil {
		return fmt.Errorf("error rendering post-load template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s); err != nil {
		return fmt.Errorf("error during post load: %s\n%w", s, err)
	}
	return nil
}

// MetaSave saves a key/value pair in the metadata table.
func (p *PostgreSQL) MetaSave(k, v string) error {
	if len(k) > 16 {
		return fmt.Errorf("metatable can only take keys that are at maximum 16 chars long")
	}
	s, err := p.renderTemplate("meta_save")
	if err != nil {
		return fmt.Errorf("error rendering meta-save template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s, k, v); err != nil {
		return fmt.Errorf("error saving %s to metadata: %w", k, err)
	}
	return nil
}

// MetaRead reads a key/value pair from the metadata table.
func (p *PostgreSQL) MetaRead(k string) (string, error) {
	rows, err := p.pool.Query(context.Background(), p.metaReadQuery, k)
	if err != nil {
		return "", fmt.Errorf("error looking for metadata key %s: %w", k, err)
	}
	v, err := pgx.CollectOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return "", fmt.Errorf("error reading for metadata key %s: %w", k, err)
	}
	return v, nil
}

// CreateExtraIndexes responsible for creating additional indexes in the database
func (p *PostgreSQL) CreateExtraIndexes(idxs []string) error {
	if err := transform.ValidateIndexes(idxs); err != nil {
		return fmt.Errorf("index name error: %w", err)
	}
	for _, idx := range idxs {
		i := ExtraIndex{
			IsRoot: !strings.Contains(idx, "."),
			Name:   fmt.Sprintf("json.%s", idx),
			Value:  idx,
		}
		p.ExtraIndexes = append(p.ExtraIndexes, i)
	}
	s, err := p.renderTemplate("extra_indexes")
	if err != nil {
		return fmt.Errorf("error rendering extra-indexes template: %w", err)
	}
	if _, err := p.pool.Exec(context.Background(), s); err != nil {
		return fmt.Errorf("expected the error to create indexe: %w", err)
	}
	slog.Info(fmt.Sprintf("%d Indexes successfully created in the table %s", len(idxs), p.CompanyTableName))
	return nil
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL(uri, schema string) (PostgreSQL, error) {
	cfg, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("could not create database config: %w", err)
	}
	cfg.MaxConns = 128
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute
	conn, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to the database: %w", err)
	}
	p := PostgreSQL{
		pool:             conn,
		uri:              uri,
		schema:           schema,
		CompanyTableName: companyTableName,
		MetaTableName:    metaTableName,
		CursorFieldName:  cursorFieldName,
		IDFieldName:      idFieldName,
		JSONFieldName:    jsonFieldName,
		KeyFieldName:     keyFieldName,
		ValueFieldName:   valueFieldName,
	}
	p.getCompanyQuery, err = p.renderTemplate("get")
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("error rendering get template: %w", err)
	}
	p.metaReadQuery, err = p.renderTemplate("meta_read")
	if err != nil {
		return PostgreSQL{}, fmt.Errorf("error rendering meta-read template: %w", err)
	}
	if err := p.pool.Ping(context.Background()); err != nil {
		return PostgreSQL{}, fmt.Errorf("could not connect to postgres: %w", err)
	}
	return p, nil
}
