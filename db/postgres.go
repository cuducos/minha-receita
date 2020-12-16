package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-pg/pg/v10"
)

// PostgreSQL database interface.
type PostgreSQL struct {
	conn   *pg.DB
	schema string
}

// Close ends the conection with the database.
func (p *PostgreSQL) Close() {
	p.conn.Close()
}

// GetCompany returns a `Company` based on a CNPJ number.
func (p *PostgreSQL) GetCompany(num string) (Company, error) {
	c, err := getCompany(p.conn, num)
	if err != nil {
		log.Output(2, fmt.Sprintf("ERROR: %v", err))
		return c, err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go queryPartners(p.conn, &wg, &c)
	go queryActivities(p.conn, &wg, &c)
	wg.Wait()
	return c, nil
}

// CreateTables creates the required database tables.
func (p *PostgreSQL) CreateTables() {
	var wg sync.WaitGroup
	src := getSources(p.schema)
	wg.Add(len(src))
	for _, s := range src {
		go createTable(p.conn, &wg, s)
	}
	wg.Wait()
}

// DropTables drops the database tables created by `CreateTables`.
func (p *PostgreSQL) DropTables() {
	var wg sync.WaitGroup
	src := getSources(p.schema)
	wg.Add(len(src))
	for _, s := range src {
		go dropTable(p.conn, &wg, s)
	}
	wg.Wait()
}

// ImportData reads data from compresed CSV and Excel files and import it.
func (p *PostgreSQL) ImportData(dir string) {
	c := make(chan error)
	src := getSources(p.schema)
	for _, s := range src {
		if s.name == "cnae" {
			go importCNAEXls(p.conn, c, s, dir)
		} else {
			go copyFrom(p.conn, c, s, dir)
		}
	}

	hasErr := false
	for i := 0; i < len(src); i++ {
		err := <-c
		if err != nil {
			hasErr = true
			log.Output(2, fmt.Sprintf("%s", err))
		}
	}
	if hasErr {
		os.Exit(1)
	}
}

// NewPostgreSQL creates a new PostgreSQL connection and ping it to make sure it works.
func NewPostgreSQL() PostgreSQL {
	u := os.Getenv("POSTGRES_URI")
	if u == "" {
		fmt.Fprintf(os.Stderr, "Please, set an environmental variable POSTGRES_URI with the credentials for the PostgreSQL database.\n")
		os.Exit(1)
	}

	s := os.Getenv("POSTGRES_SCHEMA")
	if s == "" {
		log.Output(2, "No POSTGRES_SCHEMA environment variable found, using public.")
		s = "public"
	}

	opt, err := pg.ParseURL(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse POSTGRES_URI: %v\n", err)
		os.Exit(1)
	}

	var p PostgreSQL
	p.schema = s
	p.conn = pg.Connect(opt)
	if err := p.conn.Ping(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to PostgreSQL: %v\n", err)
		os.Exit(1)
	}

	return p
}
