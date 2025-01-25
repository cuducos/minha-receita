package db

import (
	"os"
	"testing"
)

func TestMongoDB(t *testing.T) {
	id := "33683111000280"
	json := `{"qsa": [{"name": 42}, {"name": "forty-two"}], "answer": 42}`

	u := os.Getenv("MONGO_URL")
	if u == "" {
		t.Errorf("expected a posgres uri at TEST_DATABASE_URL, found nothing")
		return
	}
	mdb, err := NewMongoDB()
	if err != nil {
		t.Errorf("expected no error connecting to postgres, got %s", err)
		return
	}
	if err := mdb.DropCollection(); err != nil {
		t.Errorf("expected no error dropping the table, got %s", err)
	}
	defer func() {
		if err := mdb.DropCollection(); err != nil {
			t.Errorf("expected no error dropping the table, got %s", err)
		}
		mdb.Close()
	}()

	if err := mdb.CreateCollection(); err != nil {
		t.Errorf("expected no error creating the table, got %s", err)
	}

	if err := mdb.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a company, got %s", err)
	}
	if err := mdb.CreateCompanies([][]string{{id, json}}); err != nil {
		t.Errorf("expected no error saving a duplicated company, got %s", err)
	}

	// got, err := mdb.GetCompany("33683111000280")
	// if err != nil {
	// 	t.Errorf("expected no error getting a company, got %s", err)
	// }
	// if got != json {
	// 	t.Errorf("expected json to be %s, got %s", json, got)
	// }
	// got, err = mdb.GetCompany("33683111000280")
	// if err != nil {
	// 	t.Errorf("expected no error getting a company, got %s", err)
	// }
	// if got != json {
	// 	t.Errorf("expected json to be %s, got %s", json, got)
	// }
	// if err := mdb.MetaSaveMongo("answer", "42"); err != nil {
	// 	t.Errorf("expected no error writing to the metadata table, got %s", err)
	// }
	// metadata, err := mdb.MetaReadMongo("answer")
	// if err != nil {
	// 	t.Errorf("expected no error getting metadata, got %s", err)
	// }
	// if metadata != "42" {
	// 	t.Errorf("expected 42 as the answer, got %s", metadata)
	// }
	// if err := mdb.MetaSaveMongo("answer", "forty-two"); err != nil {
	// 	t.Errorf("expected no error re-writing to the metadata table, got %s", err)
	// }
	// metadata2, err := mdb.MetaReadMongo("answer")
	// if err != nil {
	// 	t.Errorf("expected no error getting metadata for the second time, got %s", err)
	// }
	// if metadata2 != "forty-two" {
	// 	t.Errorf("expected foruty-two as the answer, got %s", metadata2)
	// }
}
