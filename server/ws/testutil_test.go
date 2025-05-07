package ws

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
	"testing"
)

func testDB(t *testing.T) (*badger.DB, func()) {
	t.Helper()

	opts := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(nil).
		WithNumMemtables(1).            // Reduce default (5)
		WithNumLevelZeroTables(1).      // Reduce default (5)
		WithNumLevelZeroTablesStall(2). // Reduce default (10)
		WithValueLogFileSize(1 << 20)   // Reduce from default 64MB to 1MB

	db, err := badger.Open(opts)
	require.NoError(t, err)

	return db, func() {
		err := db.Close()
		require.NoError(t, err)
	}
}
