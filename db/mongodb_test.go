package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
	"go.mongodb.org/mongo-driver/bson"
)

var mongoDefaultIndexes = []string{"_id_", "id_1"}

func setUpMongo(id, c string) (*MongoDB, error) {
	u := os.Getenv("TEST_MONGODB_URL")
	if u == "" {
		return nil, fmt.Errorf("expected a mongodb uri at TEST_MONGODB_URL, found nothing")
	}
	db, err := NewMongoDB(u)
	if err != nil {
		return nil, fmt.Errorf("expected no error connecting to mongodb, got %s", err)
	}
	if err := db.Drop(); err != nil {
		return nil, fmt.Errorf("expected no error dropping the collections, got %s", err)
	}
	if err := db.Create(); err != nil {
		return nil, fmt.Errorf("expected no error creating the collections, got %s", err)
	}
	if err := db.PreLoad(); err != nil {
		return nil, fmt.Errorf("expected no error pre load on mongo, got %w", err)
	}
	if err := db.CreateCompanies([][]string{{id, c}}); err != nil {
		return nil, fmt.Errorf("expected no error saving a company to mongo, got %s", err)
	}
	if err := db.PostLoad(); err != nil {
		return nil, fmt.Errorf("expected no error post load on mongo, got %s", err)
	}
	return &db, nil
}

func listIndexesMongo(t *testing.T, db *MongoDB) []string {
	c, err := db.db.Collection(companyTableName).Indexes().List(context.Background())
	if err != nil {
		t.Errorf("expected no errors checking index list, got %s", err)
	}
	defer func() {
		if err := c.Close(context.Background()); err != nil {
			t.Errorf("expected no error closing the connection, got %s", err)
		}
	}()
	var i []string
	for c.Next(context.Background()) {
		var idx bson.M
		if err := c.Decode(&idx); err != nil {
			t.Errorf("expected no error decoding index, got %s", err)
		}
		n, ok := idx["name"].(string)
		if ok && !slices.Contains(mongoDefaultIndexes, n) {
			i = append(i, strings.TrimPrefix(n, "idx_json."))
		}
	}
	return i
}

func TestMongoCreateIndexes(t *testing.T) {
	id := "33683111000280"
	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Error("error reading company JSON file")
	}
	c := string(b)
	m, err := setUpMongo(id, c)
	if err != nil {
		t.Errorf("expected no error setting up postgres, got %s", err)
		return
	}
	defer func() {
		if err := m.Drop(); err != nil {
			t.Errorf("expected no error dropping the tables, got %s", err)
		}
		m.Close()
	}()
	i := []string{"qsa.nome_socio"}
	if err := m.CreateExtraIndexes(i); err != nil {
		t.Errorf("expected no errors running extra indexes, got %s", err)
	}
	testutils.AssertArraysHaveSameItems(t, i, listIndexesMongo(t, m))
}
