// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// SQLiteTimestampFormats is timestamp formats understood by both this module
// and SQLite.  The first format in the slice will be used when saving time
// values into the database. When parsing a string from a timestamp or datetime
// column, the formats are tried in order.
var SQLiteTimestampFormats = []string{
	// By default, store timestamps with whatever timezone they come with.
	// When parsed, they will be returned with the same timezone.
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
}

const (
	columnDate      string = "date"
	columnDatetime  string = "datetime"
	columnTimestamp string = "timestamp"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	sql.Register("sqlite3", &SQLiteDriver{})
}

// Version returns SQLite library version information.
func Version() (libVersion string, libVersionNumber int, sourceID string) {
	libVersion = sqlite3_libversion()
	libVersionNumber = sqlite3_libversion_number()
	sourceID = sqlite3_sourceid()
	return libVersion, libVersionNumber, sourceID
}

// SQLiteDriver implement driver.Driver.
type SQLiteDriver struct {
	Extensions  []string
	ConnectHook func(*SQLiteConn) error
}

type sqlite3 uintptr

// SQLiteConn implement driver.Conn.
type SQLiteConn struct {
	mu     sync.Mutex
	db     *sqlite3
	loc    *time.Location
	txlock string
}

// SQLiteTx implemen driver.Tx.
type SQLiteTx struct {
	c *SQLiteConn
}

// SQLiteStmt implement driver.Stmt.
type SQLiteStmt struct {
	mu     sync.Mutex
	c      *SQLiteConn
	s      *sqlite3_stmt
	t      string
	closed bool
	cls    bool
}

// SQLiteResult implement sql.Result.
type SQLiteResult struct {
	id      int64
	changes int64
}

// SQLiteRows implement driver.Rows.
type SQLiteRows struct {
	s        *SQLiteStmt
	nc       int
	cols     []string
	decltype []string
	cls      bool
	closed   bool
	done     chan struct{}
}

// Commit transaction.
func (tx *SQLiteTx) Commit() error {
	_, err := tx.c.exec(context.Background(), "COMMIT", nil)
	if err != nil && err.(Error).Code == SQLITE_BUSY {
		// sqlite3 will leave the transaction open in this scenario.
		// However, database/sql considers the transaction complete once we
		// return from Commit() - we must clean up to honour its semantics.
		tx.c.exec(context.Background(), "ROLLBACK", nil)
	}
	return err
}

// Rollback transaction.
func (tx *SQLiteTx) Rollback() error {
	_, err := tx.c.exec(context.Background(), "ROLLBACK", nil)
	return err
}

// AutoCommit return which currently auto commit or not.
func (c *SQLiteConn) AutoCommit() bool {
	return sqlite3_get_autocommit(*c.db) != 0
}

func (c *SQLiteConn) lastError() error {
	return lastError(c.db)
}

func lastError(db *sqlite3) error {
	rv := sqlite3_errcode(*db)
	if rv == SQLITE_OK {
		return nil
	}
	return Error{
		Code:         ErrNo(rv),
		ExtendedCode: ErrNoExtended(sqlite3_extended_errcode(*db)),
		err:          sqlite3_errmsg(*db),
	}
}

// Exec implements Execer.
func (c *SQLiteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return c.exec(context.Background(), query, list)
}

func (c *SQLiteConn) exec(ctx context.Context, query string, args []namedValue) (driver.Result, error) {
	start := 0
	for {
		s, err := c.prepare(ctx, query)
		if err != nil {
			return nil, err
		}
		var res driver.Result
		if s.(*SQLiteStmt).s != nil {
			na := s.NumInput()
			if len(args) < na {
				s.Close()
				return nil, fmt.Errorf("not enough args to execute query: want %d got %d", na, len(args))
			}
			for i := 0; i < na; i++ {
				args[i].Ordinal -= start
			}

			res, err = s.(*SQLiteStmt).exec(ctx, args[:na])
			if err != nil && err != driver.ErrSkip {
				s.Close()
				return nil, err
			}
			args = args[na:]
			start += na
		}
		tail := s.(*SQLiteStmt).t
		s.Close()
		if tail == "" {
			return res, nil
		}
		query = tail
	}
}

type namedValue struct {
	Name    string
	Ordinal int
	Value   driver.Value
}

// Query implements Queryer.
func (c *SQLiteConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return c.query(context.Background(), query, list)
}

func (c *SQLiteConn) query(ctx context.Context, query string, args []namedValue) (driver.Rows, error) {
	start := 0
	for {
		s, err := c.prepare(ctx, query)
		if err != nil {
			return nil, err
		}
		s.(*SQLiteStmt).cls = true
		na := s.NumInput()
		if len(args) < na {
			return nil, fmt.Errorf("not enough args to execute query: want %d got %d", na, len(args))
		}
		for i := 0; i < na; i++ {
			args[i].Ordinal -= start
		}
		rows, err := s.(*SQLiteStmt).query(ctx, args[:na])
		if err != nil && err != driver.ErrSkip {
			s.Close()
			return rows, err
		}
		args = args[na:]
		start += na
		tail := s.(*SQLiteStmt).t
		if tail == "" {
			return rows, nil
		}
		rows.Close()
		s.Close()
		query = tail
	}
}

// Begin transaction.
func (c *SQLiteConn) Begin() (driver.Tx, error) {
	return c.begin(context.Background())
}

func (c *SQLiteConn) begin(ctx context.Context) (driver.Tx, error) {
	if _, err := c.exec(ctx, c.txlock, nil); err != nil {
		return nil, err
	}
	return &SQLiteTx{c}, nil
}

func errorString(err Error) string {
	return sqlite3_errstr(int(err.Code))
}

// Open database and return a new connection.
// You can specify a DSN string using a URI as the filename.
//   test.db
//   file:test.db?cache=shared&mode=memory
//   :memory:
//   file::memory:
// go-sqlite3 adds the following query parameters to those used by SQLite:
//   _loc=XXX
//     Specify location of time format. It's possible to specify "auto".
//   _busy_timeout=XXX
//     Specify value for sqlite3_busy_timeout.
//   _txlock=XXX
//     Specify locking behavior for transactions.  XXX can be "immediate",
//     "deferred", "exclusive".
//   _foreign_keys=X
//     Enable or disable enforcement of foreign keys.  X can be 1 or 0.
//   _recursive_triggers=X
//     Enable or disable recursive triggers.  X can be 1 or 0.
//   _mutex=XXX
//     Specify mutex mode. XXX can be "no", "full".
func (d *SQLiteDriver) Open(dsn string) (driver.Conn, error) {
	if !sqlite3_threadsafe() {
		return nil, errors.New("sqlite library was not compiled for thread-safe operation")
	}

	var loc *time.Location
	txlock := "BEGIN"
	busyTimeout := 5000
	foreignKeys := -1
	recursiveTriggers := -1
	mutex := SQLITE_OPEN_FULLMUTEX
	pos := strings.IndexRune(dsn, '?')
	if pos >= 1 {
		params, err := url.ParseQuery(dsn[pos+1:])
		if err != nil {
			return nil, err
		}

		// _loc
		if val := params.Get("_loc"); val != "" {
			switch strings.ToLower(val) {
			case "auto":
				loc = time.Local
			default:
				loc, err = time.LoadLocation(val)
				if err != nil {
					return nil, fmt.Errorf("Invalid _loc: %v: %v", val, err)
				}
			}
		}

		// _busy_timeout
		if val := params.Get("_busy_timeout"); val != "" {
			iv, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid _busy_timeout: %v: %v", val, err)
			}
			busyTimeout = int(iv)
		}

		// _txlock
		if val := params.Get("_txlock"); val != "" {
			switch strings.ToLower(val) {
			case "immediate":
				txlock = "BEGIN IMMEDIATE"
			case "exclusive":
				txlock = "BEGIN EXCLUSIVE"
			case "deferred":
				txlock = "BEGIN"
			default:
				return nil, fmt.Errorf("Invalid _txlock: %v", val)
			}
		}

		// _foreign_keys
		if val := params.Get("_foreign_keys"); val != "" {
			switch val {
			case "1":
				foreignKeys = 1
			case "0":
				foreignKeys = 0
			default:
				return nil, fmt.Errorf("Invalid _foreign_keys: %v", val)
			}
		}

		// _recursive_triggers
		if val := params.Get("_recursive_triggers"); val != "" {
			switch val {
			case "1":
				recursiveTriggers = 1
			case "0":
				recursiveTriggers = 0
			default:
				return nil, fmt.Errorf("Invalid _recursive_triggers: %v", val)
			}
		}

		// _mutex
		if val := params.Get("_mutex"); val != "" {
			switch strings.ToLower(val) {
			case "no":
				mutex = SQLITE_OPEN_NOMUTEX
			case "full":
				mutex = SQLITE_OPEN_FULLMUTEX
			default:
				return nil, fmt.Errorf("Invalid _mutex: %v", val)
			}
		}

		if !strings.HasPrefix(dsn, "file:") {
			dsn = dsn[:pos]
		}
	}

	db := new(sqlite3)
	rv, err := sqlite3_open_v2(
		dsn,
		db,
		mutex|SQLITE_OPEN_FULLMUTEX|SQLITE_OPEN_READWRITE|SQLITE_OPEN_CREATE,
		"",
	)
	if rv != 0 {
		return nil, Error{Code: ErrNo(rv)}
	}
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("sqlite succeeded without returning a database")
	}

	rv = sqlite3_busy_timeout(*db, busyTimeout)
	if rv != SQLITE_OK {
		sqlite3_close_v2(*db)
		return nil, Error{Code: ErrNo(rv)}
	}

	exec := func(s string) error {
		rv := sqlite3_exec(*db, s, 0, 0, 0)
		if rv != SQLITE_OK {
			return lastError(db)
		}
		return err
	}
	if foreignKeys == 0 {
		if err := exec("PRAGMA foreign_keys = OFF;"); err != nil {
			sqlite3_close_v2(*db)
			return nil, err
		}
	} else if foreignKeys == 1 {
		if err := exec("PRAGMA foreign_keys = ON;"); err != nil {
			sqlite3_close_v2(*db)
			return nil, err
		}
	}
	if recursiveTriggers == 0 {
		if err := exec("PRAGMA recursive_triggers = OFF;"); err != nil {
			sqlite3_close_v2(*db)
			return nil, err
		}
	} else if recursiveTriggers == 1 {
		if err := exec("PRAGMA recursive_triggers = ON;"); err != nil {
			sqlite3_close_v2(*db)
			return nil, err
		}
	}

	conn := &SQLiteConn{db: db, loc: loc, txlock: txlock}

	if d.ConnectHook != nil {
		if err := d.ConnectHook(conn); err != nil {
			conn.Close()
			return nil, err
		}
	}
	runtime.SetFinalizer(conn, (*SQLiteConn).Close)
	return conn, nil
}

// Close the connection.
func (c *SQLiteConn) Close() error {
	rv := sqlite3_close_v2(*c.db)
	if rv != SQLITE_OK {
		return c.lastError()
	}
	c.mu.Lock()
	c.db = nil
	c.mu.Unlock()
	runtime.SetFinalizer(c, nil)
	return nil
}

func (c *SQLiteConn) dbConnOpen() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db != nil
}

// Prepare the query string. Return a new statement.
func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	return c.prepare(context.Background(), query)
}

type sqlite3_stmt uintptr

func (c *SQLiteConn) prepare(ctx context.Context, query string) (driver.Stmt, error) {
	s := new(sqlite3_stmt)
	var tail string
	rv, err := sqlite3_prepare_v2(*c.db, query, -1, s, &tail)
	if rv != SQLITE_OK {
		return nil, lastError(c.db)
	}
	if err != nil {
		return nil, err
	}

	if *s == 0 {
		return &SQLiteStmt{c: c, s: nil}, nil
	}

	ss := &SQLiteStmt{c: c, s: s, t: tail}
	runtime.SetFinalizer(ss, (*SQLiteStmt).Close)
	return ss, nil
}

// GetLimit returns the current value of a run-time limit.
// See: sqlite3_limit, http://www.sqlite.org/c3ref/limit.html
func (c *SQLiteConn) GetLimit(id int) int {
	return int(sqlite3_limit(*c.db, id, -1))
}

// SetLimit changes the value of a run-time limits.
// Then this method returns the prior value of the limit.
// See: sqlite3_limit, http://www.sqlite.org/c3ref/limit.html
func (c *SQLiteConn) SetLimit(id int, newVal int) int {
	return sqlite3_limit(*c.db, id, newVal)
}

// Close the statement.
func (s *SQLiteStmt) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if !s.c.dbConnOpen() {
		return errors.New("sqlite statement with already closed database connection")
	}
	if s != nil && s.s != nil {
		rv := sqlite3_finalize(*s.s)
		s.s = nil
		if rv != SQLITE_OK {
			return s.c.lastError()
		}
	}
	runtime.SetFinalizer(s, nil)
	return nil
}

// NumInput return a number of parameters.
func (s *SQLiteStmt) NumInput() int {
	return sqlite3_bind_parameter_count(*s.s)
}

type bindArg struct {
	n int
	v driver.Value
}

var placeHolder = []byte{0}

func (s *SQLiteStmt) bind(args []namedValue) (err error) {
	rv := sqlite3_reset(*s.s)
	if rv != SQLITE_ROW && rv != SQLITE_OK && rv != SQLITE_DONE {
		return lastError(s.c.db)
	}

	for i, v := range args {
		if v.Name != "" {
			args[i].Ordinal, err = sqlite3_bind_parameter_index(*s.s, ":"+v.Name)
			if err != nil {
				return err
			}
		}
	}

	for _, arg := range args {
		switch v := arg.Value.(type) {
		case nil:
			rv = sqlite3_bind_null(*s.s, arg.Ordinal)
		case string:
			rv = sqlite3_bind_text(*s.s, arg.Ordinal, v)
		case int64:
			rv = sqlite3_bind_int64(*s.s, arg.Ordinal, v)
		case bool:
			if v {
				rv = sqlite3_bind_int(*s.s, arg.Ordinal, 1)
			} else {
				rv = sqlite3_bind_int(*s.s, arg.Ordinal, 0)
			}
		case float64:
			rv = sqlite3_bind_double(*s.s, arg.Ordinal, v)
		case []byte:
			if v == nil {
				rv = sqlite3_bind_null(*s.s, arg.Ordinal)
			} else {
				blobLen := len(v)
				if blobLen == 0 {
					v = placeHolder
				}
				rv = sqlite3_bind_blob(*s.s, arg.Ordinal, v, blobLen)
			}
		case time.Time:
			b := v.Format(SQLiteTimestampFormats[0])
			rv = sqlite3_bind_text(*s.s, arg.Ordinal, b)
		}
		if rv != SQLITE_OK {
			return lastError(s.c.db)
		}
	}
	return nil
}

// Query the statement with arguments. Return records.
func (s *SQLiteStmt) Query(args []driver.Value) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return s.query(context.Background(), list)
}

func (s *SQLiteStmt) query(ctx context.Context, args []namedValue) (driver.Rows, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}

	rows := &SQLiteRows{
		s:        s,
		nc:       sqlite3_column_count(*s.s),
		cols:     nil,
		decltype: nil,
		cls:      s.cls,
		closed:   false,
		done:     make(chan struct{}),
	}

	if ctxdone := ctx.Done(); ctxdone != nil {
		go func(db *sqlite3) {
			select {
			case <-ctxdone:
				select {
				case <-rows.done:
				default:
					sqlite3_interrupt(*db)
					rows.Close()
				}
			case <-rows.done:
			}
		}(s.c.db)
	}

	return rows, nil
}

// LastInsertId teturn last inserted ID.
func (r *SQLiteResult) LastInsertId() (int64, error) {
	return r.id, nil
}

// RowsAffected return how many rows affected.
func (r *SQLiteResult) RowsAffected() (int64, error) {
	return r.changes, nil
}

// Exec execute the statement with arguments. Return result object.
func (s *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return s.exec(context.Background(), list)
}

func (s *SQLiteStmt) exec(ctx context.Context, args []namedValue) (driver.Result, error) {
	if err := s.bind(args); err != nil {
		sqlite3_reset(*s.s)
		sqlite3_clear_bindings(*s.s)
		return nil, err
	}

	if ctxdone := ctx.Done(); ctxdone != nil {
		done := make(chan struct{})
		defer close(done)
		go func(db *sqlite3) {
			select {
			case <-done:
			case <-ctxdone:
				select {
				case <-done:
				default:
					sqlite3_interrupt(*db)
				}
			}
		}(s.c.db)
	}

	rv := sqlite3_step(*s.s)

	var rowid, changes int64
	db := sqlite3_db_handle(*s.s)
	if db != nil {
		rowid = sqlite3_last_insert_rowid(*db)
		changes = sqlite3_changes(*db)
	}

	if rv != SQLITE_ROW && rv != SQLITE_OK && rv != SQLITE_DONE {
		err := lastError(s.c.db)
		if *s.s == 0 {
			return nil, err
		}
		sqlite3_reset(*s.s)
		sqlite3_clear_bindings(*s.s)
		return nil, err
	}

	return &SQLiteResult{id: rowid, changes: changes}, nil
}

// Close the rows.
func (rc *SQLiteRows) Close() error {
	rc.s.mu.Lock()
	if rc.s.closed || rc.closed {
		rc.s.mu.Unlock()
		return nil
	}
	rc.closed = true
	if rc.done != nil {
		close(rc.done)
	}
	if rc.cls {
		rc.s.mu.Unlock()
		return rc.s.Close()
	}
	rv := sqlite3_reset(*rc.s.s)
	if rv != SQLITE_OK {
		rc.s.mu.Unlock()
		return rc.s.c.lastError()
	}
	rc.s.mu.Unlock()
	return nil
}

// Columns return column names.
func (rc *SQLiteRows) Columns() []string {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	if rc.s.s != nil && rc.nc != len(rc.cols) {
		rc.cols = make([]string, rc.nc)
		for i := 0; i < rc.nc; i++ {
			rc.cols[i] = sqlite3_column_name(*rc.s.s, i)
		}
	}
	return rc.cols
}

func (rc *SQLiteRows) declTypes() []string {
	if rc.s.s != nil && rc.decltype == nil {
		rc.decltype = make([]string, rc.nc)
		for i := 0; i < rc.nc; i++ {
			rc.decltype[i] = strings.ToLower(sqlite3_column_decltype(*rc.s.s, i))
		}
	}
	return rc.decltype
}

// DeclTypes return column types.
func (rc *SQLiteRows) DeclTypes() []string {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	return rc.declTypes()
}

// Next move cursor to next.
func (rc *SQLiteRows) Next(dest []driver.Value) error {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	if rc.s.closed {
		return io.EOF
	}
	rv := sqlite3_step(*rc.s.s)
	if rv == SQLITE_DONE {
		return io.EOF
	}
	if rv != SQLITE_ROW {
		rv = sqlite3_reset(*rc.s.s)
		if rv != SQLITE_OK {
			return rc.s.c.lastError()
		}
		return nil
	}

	rc.declTypes()

	for i := range dest {
		switch sqlite3_column_type(*rc.s.s, i) {
		case SQLITE_INTEGER:
			val := sqlite3_column_int64(*rc.s.s, i)
			switch rc.decltype[i] {
			case columnTimestamp, columnDatetime, columnDate:
				var t time.Time
				// Assume a millisecond unix timestamp if it's 13 digits -- too
				// large to be a reasonable timestamp in seconds.
				if val > 1e12 || val < -1e12 {
					val *= int64(time.Millisecond) // convert ms to nsec
					t = time.Unix(0, val)
				} else {
					t = time.Unix(val, 0)
				}
				t = t.UTC()
				if rc.s.c.loc != nil {
					t = t.In(rc.s.c.loc)
				}
				dest[i] = t
			case "boolean":
				dest[i] = val > 0
			default:
				dest[i] = val
			}
		case SQLITE_FLOAT:
			dest[i] = sqlite3_column_double(*rc.s.s, i)
		case SQLITE_BLOB:
			p := sqlite3_column_blob(*rc.s.s, i)
			if p == 0 {
				dest[i] = nil
				continue
			}

			n := sqlite3_column_bytes(*rc.s.s, i)
			switch dest[i].(type) {
			default:
				slice := make([]byte, n)
				copy(slice, (*[1 << 30]byte)(unsafe.Pointer(p))[0:n:n])
				dest[i] = slice
			}

		case SQLITE_NULL:
			dest[i] = nil
		case SQLITE_TEXT:
			var err error
			var timeVal time.Time

			s := sqlite3_column_text(*rc.s.s, i)

			switch rc.decltype[i] {
			case columnTimestamp, columnDatetime, columnDate:
				var t time.Time
				s = strings.TrimSuffix(s, "Z")
				for _, format := range SQLiteTimestampFormats {
					if timeVal, err = time.ParseInLocation(format, s, time.UTC); err == nil {
						t = timeVal
						break
					}
				}
				if err != nil {
					// The column is a time value, so return the zero time on parse failure.
					t = time.Time{}
				}
				if rc.s.c.loc != nil {
					t = t.In(rc.s.c.loc)
				}
				dest[i] = t
			default:
				dest[i] = []byte(s)
			}
		}
	}
	return nil
}
