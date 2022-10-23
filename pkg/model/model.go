package model

import "time"

type DatabaseSchema struct {
	Name            string
	Tables          map[string]string
	TableReferences []*Table `json:"-" yaml:"-" toml:"-"`
	Generated       time.Duration
}

type Table struct {
	Name      string
	Columns   uint16
	Records   uint32
	FirstRow  int
	RowLength uint16
	FileSize  int64
	Modified  time.Time
	Fields    map[string]*Field
	Data      []map[string]interface{}
}

type Field struct {
	Name   string
	Type   string
	GoType string
	Length int
}
