package idgen

import (
	"time"
)

type IDGenerator interface {
	Generate() (int64, error)
	GenerateString() (string, error)
	Clone() IDGenerator
	Parse(id int64) ParsedID
}

// ParseID 解析ID的组成部分
type ParsedID struct {
	Timestamp time.Time
	MachineID int64
	Sequence  int64
}
