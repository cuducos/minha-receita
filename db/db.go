// Package db implements the high level API for a database interface. The lines
// in this file should be agnostic in terms of the database provider.
//
// Files such as `postgres.go` implements a specific database provider.
package db

const (
	tableName     = "cnpj"
	idFieldName   = "id"
	jsonFieldName = "json"
)

// Database interface to Minha Receita.
type Database interface {
	CreateTable() error
	DropTable() error
	ImportData(string) error
	GetCompany(string) (string, error)
	Close()
}
