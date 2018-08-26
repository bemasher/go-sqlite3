package sqlite3

import (
	"bytes"
	"strings"
	"syscall"
	"unsafe"
)

//go:generate go run mksyscall/mksyscall_windows.go -systemdll=false -output zsqlite3_gen.go sqlite3_gen.go
// //go:generate go run mksyscall/mksyscall_windows.go -trace=true -systemdll=false -output zsqlite3_gen.go sqlite3_gen.go

const sliceLen = 1<<31 - 1

func UintptrToString(p uintptr) string {
	if p == 0 {
		return ""
	}

	buf := (*[sliceLen]byte)(unsafe.Pointer(p))
	n := bytes.IndexByte((*buf)[:], 0)
	s := make([]byte, n)
	copy(s, buf[0:n])

	return string(s)
}

func BytePtrToString(p *byte) string {
	buf := (*[sliceLen]byte)(unsafe.Pointer(p))
	n := bytes.IndexByte((*buf)[:], 0)
	s := make([]byte, n)
	copy(s, buf[0:n])

	return string(s)
}

//sys sqlite3_libversion() (version dllString) = sqlite3.sqlite3_libversion
//sys sqlite3_libversion_number() (version int) = sqlite3.sqlite3_libversion_number
//sys sqlite3_sourceid() (sourceidPtr dllString) = sqlite3.sqlite3_sourceid
//sys sqlite3_threadsafe() (isThreadSafe bool) = sqlite3.sqlite3_threadsafe

//sys sqlite3_busy_timeout(db sqlite3, ms int) (ret int) = sqlite3.sqlite3_busy_timeout

// //sys sqlite3_exec(db sqlite3, sql string, callback uintptr, args uintptr, errMsg *string) (ret int, err error) = sqlite3.sqlite3_exec
//sys zsqlite3_exec() = sqlite3.sqlite3_exec
func sqlite3_exec(db sqlite3, sql string, callback uintptr, args uintptr, errMsg uintptr) (ret int) {
	_p0, err := syscall.BytePtrFromString(sql)
	if err != nil {
		return
	}

	r0, _, _ := procsqlite3_exec.Call(
		uintptr(db),
		uintptr(unsafe.Pointer(_p0)),
		callback,
		args,
		errMsg,
	)

	ret = int(r0)
	return
}

//sys sqlite3_errcode(db sqlite3) (errcode int) = sqlite3.sqlite3_errcode
//sys sqlite3_extended_errcode(db sqlite3) (errcode int) = sqlite3.sqlite3_extended_errcode
//sys sqlite3_errmsg(db sqlite3) (msg dllString) = sqlite3.sqlite3_errmsg
//sys sqlite3_errstr(err int) (errStrPtr dllString) = sqlite3.sqlite3_errstr

// //sys sqlite3_open_v2(filename string, db sqlite3, flags int, vfs string) (ret int, err error) = sqlite3.sqlite3_open_v2
//sys zsqlite3_open_v2() = sqlite3.sqlite3_open_v2
func sqlite3_open_v2(filename string, db *sqlite3, flags int, vfs string) (ret int, err error) {
	// print("SYSCALL: sqlite3_open_v2(", "filename=", filename, ")\n")
	var (
		_p0 *byte
		_u0 uintptr
	)
	if len(filename) > 0 {
		_p0, err = syscall.BytePtrFromString(filename)
		if err != nil {
			return
		}
		_u0 = uintptr(unsafe.Pointer(_p0))
	}

	var (
		_p1 *byte
		_u1 uintptr
	)
	if len(vfs) > 0 {
		_p1, err = syscall.BytePtrFromString(vfs)
		if err != nil {
			return
		}
		_u1 = uintptr(unsafe.Pointer(_p1))
	}

	r0, _, _ := procsqlite3_open_v2.Call(
		_u0,
		uintptr(unsafe.Pointer(db)),
		uintptr(flags|SQLITE_OPEN_URI),
		_u1,
	)
	ret = int(r0)
	return
}

//sys sqlite3_close_v2(db sqlite3) (ret int) = sqlite3.sqlite3_close_v2

// //sys sqlite3_backup_init(dstdb sqlite3, dst string, srcdb sqlite3, src string) (b sqlite3_backup) = sqlite3.sqlite3_backup_init
//sys zsqlite3_backup_init() = sqlite3.sqlite3_backup_init
func sqlite3_backup_init(dstdb sqlite3, dst string, srcdb sqlite3, src string) (b sqlite3_backup, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(dst)
	if err != nil {
		return
	}
	var _p1 *byte
	_p1, err = syscall.BytePtrFromString(src)
	if err != nil {
		return
	}

	r0, _, _ := procsqlite3_backup_init.Call(
		uintptr(dstdb),
		uintptr(unsafe.Pointer(_p0)),
		uintptr(srcdb),
		uintptr(unsafe.Pointer(_p1)),
	)

	b = sqlite3_backup(r0)

	return
}

//sys sqlite3_backup_pagecount(b sqlite3_backup) (ret int) = sqlite3.sqlite3_backup_pagecount
//sys sqlite3_backup_remaining(b sqlite3_backup) (ret int) = sqlite3.sqlite3_backup_remaining
//sys sqlite3_backup_step(b sqlite3_backup, page int) (ret int) = sqlite3.sqlite3_backup_step
//sys sqlite3_backup_finish(b sqlite3_backup) (ret int) = sqlite3.sqlite3_backup_finish

//sys sqlite3_get_autocommit(db sqlite3) (ret int) = sqlite3.sqlite3_get_autocommit

//sys zsqlite3_prepare_v2() = sqlite3.sqlite3_prepare_v2
func sqlite3_prepare_v2(db sqlite3, sql string, n int, s *sqlite3_stmt, tail *string) (ret int, err error) {
	var pSql *byte
	pSql, err = syscall.BytePtrFromString(sql)
	if err != nil {
		return 0, err
	}

	pTail := new(uintptr)

	r0, _, _ := procsqlite3_prepare_v2.Call(
		uintptr(unsafe.Pointer(db)),
		uintptr(unsafe.Pointer(pSql)),
		uintptr(SQLITE_TRANSIENT),
		uintptr(unsafe.Pointer(s)),
		uintptr(unsafe.Pointer(pTail)),
	)

	*tail = sql[*pTail-uintptr(unsafe.Pointer(pSql)):]
	*tail = strings.TrimSpace(*tail)

	return int(r0), nil
}

//sys sqlite3_limit(db sqlite3, id int, newVal int) (ret int) = sqlite3.sqlite3_limit
//sys sqlite3_finalize(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_finalize
//sys sqlite3_bind_parameter_count(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_bind_parameter_count

//sys sqlite3_reset(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_reset
// //sys sqlite3_bind_parameter_index(stmt sqlite3_stmt, name string) (ret int, err error) = sqlite3.sqlite3_bind_parameter_index
//sys zsqlite3_bind_parameter_index() = sqlite3.sqlite3_bind_parameter_index
func sqlite3_bind_parameter_index(stmt sqlite3_stmt, name string) (ret int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(name)
	if err != nil {
		return
	}

	var r0 uintptr
	r0, _, err = procsqlite3_bind_parameter_index.Call(uintptr(stmt), uintptr(unsafe.Pointer(_p0)))
	ret = int(r0)
	return
}

//sys sqlite3_bind_null(stmt sqlite3_stmt, ordinal int) (ret int) = sqlite3.sqlite3_bind_null

//sys zsqlite3_bind_blob() = sqlite3.sqlite3_bind_blob
func sqlite3_bind_blob(stmt sqlite3_stmt, ordinal int, blob []byte, l int) (ret int) {
	var pBlob uintptr
	pBlob = uintptr(unsafe.Pointer(&blob[0]))

	r0, _, _ := procsqlite3_bind_blob.Call(
		uintptr(unsafe.Pointer(stmt)),
		uintptr(ordinal),
		pBlob,
		uintptr(l),
		uintptr(SQLITE_TRANSIENT),
	)
	return int(r0)
}

//sys zsqlite3_bind_text() = sqlite3.sqlite3_bind_text
func sqlite3_bind_text(stmt sqlite3_stmt, ordinal int, text string) (ret int) {
	var pText uintptr
	if len(text) > 0 {
		pText = uintptr(unsafe.Pointer(&[]byte(text)[0]))
	}

	r0, _, _ := procsqlite3_bind_text.Call(
		uintptr(unsafe.Pointer(stmt)),
		uintptr(ordinal),
		pText,
		uintptr(len(text)),
		uintptr(SQLITE_TRANSIENT),
	)
	return int(r0)
}

//sys sqlite3_bind_value(stmt sqlite3_stmt, ordinal int, val uintptr) (ret int) = sqlite3.sqlite3_bind_value

//sys zsqlite3_bind_double() = sqlite3.sqlite3_bind_double
func sqlite3_bind_double(stmt sqlite3_stmt, ordinal int, val float64) (ret int) {
	r0, _, _ := procsqlite3_bind_double.Call(uintptr(stmt), uintptr(ordinal), *(*uintptr)(unsafe.Pointer(&val)))
	ret = int(r0)
	return
}

//sys sqlite3_bind_int(stmt sqlite3_stmt, ordinal int, val int) (ret int) = sqlite3.sqlite3_bind_int
//sys sqlite3_bind_int64(stmt sqlite3_stmt, ordinal int, val int64) (ret int) = sqlite3.sqlite3_bind_int64

//sys sqlite3_column_count(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_column_count

//sys sqlite3_interrupt(db sqlite3) = sqlite3.sqlite3_interrupt

//sys sqlite3_clear_bindings(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_clear_bindings

//sys sqlite3_step(stmt sqlite3_stmt) (ret int) = sqlite3.sqlite3_step

// //sys sqlite3_db_handle(stmt sqlite3_stmt) (db *sqlite3) = sqlite3.sqlite3_db_handle
//sys zsqlite3_db_handle() = sqlite3.sqlite3_db_handle
func sqlite3_db_handle(stmt sqlite3_stmt) (db *sqlite3) {
	r0, _, _ := procsqlite3_db_handle.Call(uintptr(stmt))
	return (*sqlite3)(unsafe.Pointer(&r0))
}

//sys sqlite3_last_insert_rowid(db sqlite3) (rowid int64) = sqlite3.sqlite3_last_insert_rowid
//sys sqlite3_changes(db sqlite3) (changes int64) = sqlite3.sqlite3_changes

//sys sqlite3_column_name(stmt sqlite3_stmt, idx int) (name dllString) = sqlite3.sqlite3_column_name
//sys sqlite3_column_decltype(stmt sqlite3_stmt, idx int) (name dllString) = sqlite3.sqlite3_column_decltype

//sys sqlite3_column_bytes(stmt sqlite3_stmt, idx int) (ret int) = sqlite3.sqlite3_column_bytes
//sys zsqlite3_column_blob() = sqlite3.sqlite3_column_blob
func sqlite3_column_blob(stmt sqlite3_stmt, idx int) uintptr {
	bytePtr, _, _ := procsqlite3_column_blob.Call(
		uintptr(unsafe.Pointer(stmt)),
		uintptr(idx),
	)

	return uintptr(bytePtr)
}

//sys sqlite3_value_dup(unprot_val uintptr) (prot_val uintptr) = sqlite3.sqlite3_value_dup

// //sys sqlite3_column_double(stmt sqlite3_stmt, idx int) (ret float64) = sqlite3.sqlite3_column_double
//sys zsqlite3_column_double() = sqlite3.sqlite3_column_double
func sqlite3_column_double(stmt sqlite3_stmt, idx int) float64 {
	unprot_val, _, _ := procsqlite3_column_value.Call(
		uintptr(unsafe.Pointer(stmt)),
		uintptr(idx),
	)

	ptr, _, _ := procsqlite3_value_dup.Call(unprot_val)

	return *(*float64)(unsafe.Pointer(ptr))
}

//sys sqlite3_column_int64(stmt sqlite3_stmt, idx int) (ret int64) = sqlite3.sqlite3_column_int64

//sys zsqlite3_column_text() = sqlite3.sqlite3_column_text
func sqlite3_column_text(stmt sqlite3_stmt, idx int) string {
	bytePtr, _, _ := procsqlite3_column_text.Call(
		uintptr(unsafe.Pointer(stmt)),
		uintptr(idx),
	)

	n := sqlite3_column_bytes(stmt, idx)
	slice := make([]byte, n)
	copy(slice, (*[sliceLen]byte)(unsafe.Pointer(bytePtr))[:n:n])

	return string(slice)
}

//sys sqlite3_column_type(stmt sqlite3_stmt, idx int) (ret int) = sqlite3.sqlite3_column_type

//sys sqlite3_column_value(stmt sqlite3_stmt, idx int) (ret uintptr) = sqlite3.sqlite3_column_value
