package db

import (
	"context"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cuducos/minha-receita/testutils"
	"go.mongodb.org/mongo-driver/bson"
)

var mongoDefaultIndexes = []string{"_id_", "id_1"}

func listIndexesMongo(t *testing.T, db *MongoDB) []string {
	c, err := db.db.Collection(companyTableName).Indexes().List(context.Background())
	if err != nil {
		t.Errorf("expected no errors checking index list, got %s", err)
	}
	defer c.Close(context.Background())
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

func TestMongoDB(t *testing.T) {
	id := "19131243000197"
	b, err := os.ReadFile(filepath.Join("..", "testdata", "response.json"))
	if err != nil {
		t.Error("error reading company JSON file")
	}
	c := string(b)

	// ignore date conversions form string to date when writing to the database
	c = strings.ReplaceAll(c, "2013-10-03", "0001-01-01")
	c = strings.ReplaceAll(c, "2024-02-27", "0001-01-01")

	u := os.Getenv("TEST_MONGODB_URL")
	if u == "" {
		t.Errorf("expected a mongodb uri at TEST_MONGODB_URL, found nothing")
		return
	}
	db, err := NewMongoDB(u)
	if err != nil {
		t.Errorf("expected no error connecting to mongodb, got %s", err)
		return
	}
	if err := db.Drop(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
	defer func() {
		if err := db.Drop(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		db.Close()
	}()

	if err := db.Create(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}

	if err := db.CreateCompanies([][]string{{id, c}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := db.CreateCompanies([][]string{{id, c}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}
	if err := db.PostLoad(); err != nil {
		t.Errorf("expected no error post load, got %s", err)
	}
	got, err := db.GetCompany("19131243000197")
	if err != nil {
		t.Errorf("expected no error getting a company, got %s", err)
	}
	assertCompaniesAreEqual(t, got, c)
	if err := db.MetaSave("answer", "42"); err != nil {
		t.Errorf("expected no error writing to the metadata table, got %s", err)
	}
	var q Query
	q.Limit = 1
	q.UF = []string{"RJ"}
	sr, err := db.Search(context.Background(), &q)
	if err != nil {
		t.Errorf("expected no error querying %#v, got %s", q, err)
	}
	var p page
	if err := json.Unmarshal([]byte(sr), &p); err != nil {
		t.Errorf("expected error deserializing JSON %s, got %s", sr, err)
	}
	if len(p.Data) != 0 {
		t.Errorf("expected error no result, got %#v", p)
	}
	q.UF = []string{"SP"}
	sr, err = db.Search(context.Background(), &q)
	if err != nil {
		t.Errorf("expected no error querying %#v, got %s", q, err)
	}
	if err := json.Unmarshal([]byte(sr), &p); err != nil {
		t.Errorf("expected no error deserializing JSON %s, got %s", sr, err)
	}
	if len(p.Data) != 1 {
		t.Errorf("expected one result, got %d", len(p.Data))
	}
	if p.Data[0].UF != "SP" {
		t.Errorf("expected query result to be from SP, got %s", p.Data[0].UF)
	}
	metadata, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata, got %s", err)
	}
	if metadata != "42" {
		t.Errorf("expected 42 as the answer, got %s", metadata)
	}
	if err := db.MetaSave("answer", "forty-two"); err != nil {
		t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	}
	metadata2, err := db.MetaRead("answer")
	if err != nil {
		t.Errorf("expected no error getting metadata for the second time, got %s", err)
	}
	if metadata2 != "forty-two" {
		t.Errorf("expected foruty-two as the answer, got %s", metadata2)
	}
	if err := db.CreateExtraIndexes([]string{"teste.index1"}); err == nil {
		t.Error("expected errors running extra indexes, got nil")
	}
	i := []string{"qsa.nome_socio"}
	if err := db.CreateExtraIndexes(i); err != nil {
		t.Errorf("expected no errors running extra indexes, got %s", err)
	}
	testutils.AssertArraysHaveSameItems(t, i, listIndexesMongo(t, &db))
}
