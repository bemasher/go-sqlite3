package main

import (
	"database/sql"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestPRAGMA(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, pragma := range []string{
		`PRAGMA foreign_keys;`,
		`PRAGMA recursive_triggers;`,
	} {
		r := db.QueryRow(pragma)
		var val interface{}
		err := r.Scan(&val)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v\n", val)
	}

	for _, pragma := range []string{
		`PRAGMA foreign_keys = OFF;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA recursive_triggers = OFF;`,
		`PRAGMA recursive_triggers = ON;`,
	} {
		_, err := db.Exec(pragma)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, pragma := range []string{
		`PRAGMA foreign_keys;`,
		`PRAGMA recursive_triggers;`,
	} {
		r := db.QueryRow(pragma)
		var val interface{}
		err := r.Scan(&val)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v\n", val)
	}
}
