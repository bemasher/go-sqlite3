package sqlite3

const (
	SQLITE_DELETE = 9  // Table Name      NULL
	SQLITE_INSERT = 18 // Table Name      NULL
	SQLITE_UPDATE = 23 // Table Name      Column Name
)

// Fundamental Datatypes
const (
	SQLITE_INTEGER = 1
	SQLITE_FLOAT   = 2
	SQLITE_TEXT    = 3
	SQLITE_BLOB    = 4
	SQLITE_NULL    = 5
)

// Result Codes
const (
	SQLITE_OK         = 0   // Successful result
	SQLITE_ERROR      = 1   // Generic error
	SQLITE_INTERNAL   = 2   // Internal logic error in SQLite
	SQLITE_PERM       = 3   // Access permission denied
	SQLITE_ABORT      = 4   // Callback routine requested an abort
	SQLITE_BUSY       = 5   // The database file is locked
	SQLITE_LOCKED     = 6   // A table in the database is locked
	SQLITE_NOMEM      = 7   // A malloc() failed
	SQLITE_READONLY   = 8   // Attempt to write a readonly database
	SQLITE_INTERRUPT  = 9   // Operation terminated by sqlite3_interrupt(
	SQLITE_IOERR      = 10  // Some kind of disk I/O error occurred
	SQLITE_CORRUPT    = 11  // The database disk image is malformed
	SQLITE_NOTFOUND   = 12  // Unknown opcode in sqlite3_file_control()
	SQLITE_FULL       = 13  // Insertion failed because database is full
	SQLITE_CANTOPEN   = 14  // Unable to open the database file
	SQLITE_PROTOCOL   = 15  // Database lock protocol error
	SQLITE_EMPTY      = 16  // Internal use only
	SQLITE_SCHEMA     = 17  // The database schema changed
	SQLITE_TOOBIG     = 18  // String or BLOB exceeds size limit
	SQLITE_CONSTRAINT = 19  // Abort due to constraint violation
	SQLITE_MISMATCH   = 20  // Data type mismatch
	SQLITE_MISUSE     = 21  // Library used incorrectly
	SQLITE_NOLFS      = 22  // Uses OS features not supported on host
	SQLITE_AUTH       = 23  // Authorization denied
	SQLITE_FORMAT     = 24  // Not used
	SQLITE_RANGE      = 25  // 2nd parameter to sqlite3_bind out of range
	SQLITE_NOTADB     = 26  // File opened that is not a database file
	SQLITE_NOTICE     = 27  // Notifications from sqlite3_log()
	SQLITE_WARNING    = 28  // Warnings from sqlite3_log()
	SQLITE_ROW        = 100 // sqlite3_step() has another row ready
	SQLITE_DONE       = 101 // sqlite3_step() has finished executing
)

// Text Encodings
const (
	SQLITE_UTF8          = 1 // IMP: R-37514-35566
	SQLITE_UTF16LE       = 2 // IMP: R-03371-37637
	SQLITE_UTF16BE       = 3 // IMP: R-51971-34154
	SQLITE_UTF16         = 4 // Use native byte order
	SQLITE_ANY           = 5 // Deprecated
	SQLITE_UTF16_ALIGNED = 8 // sqlite3_create_collation only
)

// Flags For File Open Operations
const (
	SQLITE_OPEN_READONLY       = 0x00000001 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_READWRITE      = 0x00000002 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_CREATE         = 0x00000004 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_DELETEONCLOSE  = 0x00000008 // VFS only
	SQLITE_OPEN_EXCLUSIVE      = 0x00000010 // VFS only
	SQLITE_OPEN_AUTOPROXY      = 0x00000020 // VFS only
	SQLITE_OPEN_URI            = 0x00000040 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_MEMORY         = 0x00000080 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_MAIN_DB        = 0x00000100 // VFS only
	SQLITE_OPEN_TEMP_DB        = 0x00000200 // VFS only
	SQLITE_OPEN_TRANSIENT_DB   = 0x00000400 // VFS only
	SQLITE_OPEN_MAIN_JOURNAL   = 0x00000800 // VFS only
	SQLITE_OPEN_TEMP_JOURNAL   = 0x00001000 // VFS only
	SQLITE_OPEN_SUBJOURNAL     = 0x00002000 // VFS only
	SQLITE_OPEN_MASTER_JOURNAL = 0x00004000 // VFS only
	SQLITE_OPEN_NOMUTEX        = 0x00008000 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_FULLMUTEX      = 0x00010000 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_SHAREDCACHE    = 0x00020000 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_PRIVATECACHE   = 0x00040000 // Ok for sqlite3_open_v2()
	SQLITE_OPEN_WAL            = 0x00080000 // VFS only
)

const SQLITE_TRANSIENT = 18446744073709551615

// Run-Time Limit Categories
const (
	SQLITE_LIMIT_LENGTH              = 0
	SQLITE_LIMIT_SQL_LENGTH          = 1
	SQLITE_LIMIT_COLUMN              = 2
	SQLITE_LIMIT_EXPR_DEPTH          = 3
	SQLITE_LIMIT_COMPOUND_SELECT     = 4
	SQLITE_LIMIT_VDBE_OP             = 5
	SQLITE_LIMIT_FUNCTION_ARG        = 6
	SQLITE_LIMIT_ATTACHED            = 7
	SQLITE_LIMIT_LIKE_PATTERN_LENGTH = 8
	SQLITE_LIMIT_VARIABLE_NUMBER     = 9
	SQLITE_LIMIT_TRIGGER_DEPTH       = 10
	SQLITE_LIMIT_WORKER_THREADS      = 11
)
