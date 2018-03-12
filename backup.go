// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"runtime"
)

type sqlite3_backup uintptr

// SQLiteBackup implement interface of Backup.
type SQLiteBackup struct {
	b *sqlite3_backup
}

// Backup make backup from src to dest.
func (c *SQLiteConn) Backup(dst string, conn *SQLiteConn, src string) (*SQLiteBackup, error) {
	if b, err := sqlite3_backup_init(*c.db, dst, *conn.db, src); b != 0 || err != nil {
		bb := &SQLiteBackup{b: &b}
		runtime.SetFinalizer(bb, (*SQLiteBackup).Finish)
		return bb, err
	}
	return nil, c.lastError()
}

// Step to backs up for one step. Calls the underlying `sqlite3_backup_step`
// function.  This function returns a boolean indicating if the backup is done
// and an error signalling any other error. Done is returned if the underlying
// C function returns SQLITE_DONE (Code 101)
func (b *SQLiteBackup) Step(p int) (bool, error) {
	ret := sqlite3_backup_step(*b.b, p)
	if ret == SQLITE_DONE {
		return true, nil
	} else if ret != 0 && ret != SQLITE_LOCKED && ret != SQLITE_BUSY {
		return false, Error{Code: ErrNo(ret)}
	}
	return false, nil
}

// Remaining return whether have the rest for backup.
func (b *SQLiteBackup) Remaining() int {
	return sqlite3_backup_remaining(*b.b)
}

// PageCount return count of pages.
func (b *SQLiteBackup) PageCount() int {
	return sqlite3_backup_pagecount(*b.b)
}

// Finish close backup.
func (b *SQLiteBackup) Finish() error {
	return b.Close()
}

// Close close backup.
func (b *SQLiteBackup) Close() error {
	ret := sqlite3_backup_finish(*b.b)

	// sqlite3_backup_finish() never fails, it just returns the
	// error code from previous operations, so clean up before
	// checking and returning an error
	b.b = nil
	runtime.SetFinalizer(b, nil)

	if ret != 0 {
		return Error{Code: ErrNo(ret)}
	}
	return nil
}
