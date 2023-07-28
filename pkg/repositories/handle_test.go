package repositories

import (
	"context"
	"path"
	"testing"

	"github.com/flyteorg/flytestdlib/database"
	"github.com/stretchr/testify/assert"
)

func TestNewDBHandle(t *testing.T) {
	t.Run("missing DB Config", func(t *testing.T) {
		_, err := NewDBHandle(context.TODO(), database.DbConfig{}, nil)
		assert.Error(t, err)
	})

	t.Run("sqlite config", func(t *testing.T) {
		dbFile := path.Join(t.TempDir(), "admin.db")
		dbHandle, err := NewDBHandle(context.TODO(), database.DbConfig{SQLite: database.SQLiteConfig{File: dbFile}}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, dbHandle)
	})
}
