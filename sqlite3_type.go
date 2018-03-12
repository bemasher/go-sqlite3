package sqlite3

import (
	"reflect"
	"time"
)

// ColumnTypeDatabaseTypeName implement RowsColumnTypeDatabaseTypeName.
func (rc *SQLiteRows) ColumnTypeDatabaseTypeName(i int) string {
	return sqlite3_column_decltype(*rc.s.s, i)
}

/*
func (rc *SQLiteRows) ColumnTypeLength(index int) (length int64, ok bool) {
	return 0, false
}

func (rc *SQLiteRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, false
}
*/

// ColumnTypeNullable implement RowsColumnTypeNullable.
func (rc *SQLiteRows) ColumnTypeNullable(i int) (nullable, ok bool) {
	return true, true
}

// ColumnTypeScanType implement RowsColumnTypeScanType.
func (rc *SQLiteRows) ColumnTypeScanType(i int) reflect.Type {
	switch sqlite3_column_type(*rc.s.s, i) {
	case SQLITE_INTEGER:
		switch sqlite3_column_decltype(*rc.s.s, i) {
		case "timestamp", "datetime", "date":
			return reflect.TypeOf(time.Time{})
		case "boolean":
			return reflect.TypeOf(false)
		}
		return reflect.TypeOf(int64(0))
	case SQLITE_FLOAT:
		return reflect.TypeOf(float64(0))
	case SQLITE_BLOB:
		return reflect.SliceOf(reflect.TypeOf(byte(0)))
	case SQLITE_NULL:
		return reflect.TypeOf(nil)
	case SQLITE_TEXT:
		return reflect.TypeOf("")
	}
	return reflect.SliceOf(reflect.TypeOf(byte(0)))
}
