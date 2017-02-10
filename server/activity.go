package server

import (
	"time"
)

type Activity interface {
	ActivityTime() time.Time
	ActivityKey() uint32
}
