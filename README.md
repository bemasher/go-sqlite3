# go-sqlite3

[![GoDoc Reference](https://godoc.org/github.com/bemasher/go-sqlite3?status.svg)](http://godoc.org/github.com/mattn/go-sqlite3)

## Description

Sqlite3 driver conforming to the built-in database/sql interface.

This is a modified version of mattn's sqlite3 driver intended for windows users that do not wish to have a cgo dependency that makes use of a sqlite3 dll instead.

## Version

This package is currently built and tested against: `sqlite 3.23.1`

## Installation

This package can be installed with the go get command:

    go get github.com/bemasher/go-sqlite3

Also be sure to download the pre-compiled dll from the sqlite website and place it in your path: https://www.sqlite.org/download.html

## License


MIT: http://mattn.mit-license.org/2012

## Author

 * Yasuhiro Matsumoto (mattn)
 * Douglas Hall (bemasher)