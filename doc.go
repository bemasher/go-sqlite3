/*
Package sqlite3 provides interface to SQLite3 databases.

This works as a driver for database/sql.

Installation

	go get github.com/bemasher/go-sqlite3

Also be sure to download the precompiled dll and add it to a location in your path:

	https://www.sqlite.org/download.html

Supported Types

Currently, go-sqlite3 supports the following data types.

	+------------------------------+
	|go        | sqlite3           |
	|----------|-------------------|
	|nil       | null              |
	|int       | integer           |
	|int64     | integer           |
	|float64   | float             |
	|bool      | integer           |
	|[]byte    | blob              |
	|string    | text              |
	|time.Time | timestamp/datetime|
	+------------------------------+

Connection Hook

You can hook and inject your code when the connection is established. database/sql
doesn't provide a way to get native go-sqlite3 interfaces. So if you want,
you need to set ConnectHook and get the SQLiteConn.

	sql.Register("sqlite3_with_hook_example",
		&sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					sqlite3conn = append(sqlite3conn, conn)
					return nil
				},
		},
	)
*/
package sqlite3
