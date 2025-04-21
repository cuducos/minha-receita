package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cuducos/minha-receita/db"
)

var (
	databaseURI    string
	postgresSchema string
)

type database interface {
	Create() error
	Drop() error
	Close()
	// transform
	PreLoad() error
	CreateCompanies([][]string) error
	PostLoad() error
	MetaSave(string, string) error
        // extra indexes
	ExtraIndexes(idxs []string) error
	// api
	GetCompany(string) (string, error)
	MetaRead(string) (string, error)
}

func loadDatabase() (database, error) {
	var u string
	if databaseURI != "" {
		u = databaseURI
	} else {
		u = os.Getenv("DATABASE_URL")
	}
	if u == "" {
		return nil, fmt.Errorf("could not find a database URI, set the DATABASE_URL environment variable with the credentials for a database")
	}
	if strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://") {
		db, err := db.NewPostgreSQL(u, postgresSchema)
		return &db, err
	}
	if strings.HasPrefix(u, "mongodb://") {
		db, err := db.NewMongoDB(u)
		return &db, err
	}
	return nil, fmt.Errorf("database uri does not seem to be a valid Postgres or MongoDB URI")
}
