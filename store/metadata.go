package store

import "time"

type Metadata struct {
	filename string
	atime    time.Time
	size     int64
}