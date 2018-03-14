package main

import (
	"database/sql"
	"math"
	"strconv"
	"testing"

	_ "github.com/bemasher/go-sqlite3"
)

func fatalNil(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestDouble(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	fatalNil(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE Foo (Val REAL);")
	fatalNil(t, err)

	ins, err := db.Prepare("INSERT INTO Foo VALUES (?);")
	fatalNil(t, err)

	for _, val := range []float64{math.Pi, math.Phi} {
		_, err := ins.Exec(val)
		fatalNil(t, err)
	}

	for _, val := range []float64{math.Pi, math.Phi} {
		_, err := ins.Exec(strconv.FormatFloat(val, 'f', 15, 64))
		if err != nil {
			t.Fatal(err)
		}
	}

	rows, err := db.Query("SELECT * FROM Foo;")
	fatalNil(t, err)

	for rows.Next() {
		var (
			floatVal  sql.NullFloat64
			stringVal sql.NullString
		)
		err = rows.Scan(&floatVal)
		if err != nil {
			t.Fatal(err)
		}

		err = rows.Scan(&stringVal)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Val: %s %0.15f\n", stringVal.String, floatVal.Float64)
	}
}

func TestLastInserIDRowsAffected(t *testing.T) {
	type testCase struct {
		sql          string
		lastInsertId int64
		rowsAffected int64
	}

	db, err := sql.Open("sqlite3", ":memory:")
	fatalNil(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE Foo (Val INTEGER);")
	fatalNil(t, err)

	for _, tc := range []testCase{
		{"INSERT INTO Foo VALUES (1), (2), (3);", 3, 3},
		{"DELETE FROM Foo;", 3, 3},
	} {
		r, err := db.Exec(tc.sql)
		fatalNil(t, err)
		li, _ := r.LastInsertId()
		ra, _ := r.RowsAffected()
		if li != tc.lastInsertId || ra != tc.rowsAffected {
			t.Fatalf("RowsAffected: % 3d != % 3d LastInsertId: % 3d != % 3d\n", ra, tc.rowsAffected, li, tc.lastInsertId)
		}
		t.Logf("RowsAffected: % 3d LastInsertId: % 3d\n", ra, li)
	}
}
