package gormimpl

import (
	"time"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
)

// Common metrics for DB CRUD operations
type gormMetrics struct {
	Scope          promutils.Scope
	CreateDuration labeled.StopWatch
	GetDuration    labeled.StopWatch
	ListDuration   labeled.StopWatch
}

func newGormMetrics(scope promutils.Scope) gormMetrics {
	return gormMetrics{
		Scope: scope,
		CreateDuration: labeled.NewStopWatch(
			"create", "Duration for creating a new entity", time.Millisecond, scope),
		GetDuration: labeled.NewStopWatch(
			"get", "Duration for retrieving an entity ", time.Millisecond, scope),
		ListDuration: labeled.NewStopWatch(
			"list", "Duration for listing entities ", time.Millisecond, scope),
	}
}
