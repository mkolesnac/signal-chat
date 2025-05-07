package database

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabase_OpenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("creates db files in user directory", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)

		// Act
		err := db.Open("123")

		// Assert
		assert.NoError(t, err)
		got := filepath.Join(tf, "123")
		assert.DirExists(t, got, "database directory should have been created")

		err = db.Write("test", []byte("test"))
		assert.NoError(t, err, "should be able to write to the opened database")
	})
}

func TestDatabase_WriteValue(t *testing.T) {
	t.Run("panics when database not opened", func(t *testing.T) {
		// Arrange
		db := NewDatabase()
		// Act & Assert
		assert.Panics(t, func() { _ = db.Write("123", []byte("test")) })
	})
}

func TestDatabase_WriteValueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("overwrites existing value successfully", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		err = db.Write("key1", []byte("initial"))

		// Assert
		assert.NoError(t, err)
		got, err := db.Read("key1")
		assert.NoError(t, err)
		assert.Equal(t, "initial", string(got))

		err = db.Write("key1", []byte("updated"))
		assert.NoError(t, err)

		got, err = db.Read("key1")
		assert.NoError(t, err)
		assert.Equal(t, "updated", string(got))
	})
}

func TestDatabase_ReadValue(t *testing.T) {
	t.Run("panics when database not opened", func(t *testing.T) {
		// Arrange
		db := NewDatabase()
		// Act&Assert
		assert.Panics(t, func() { _, _ = db.Read("123") })
	})
}

func TestDatabase_ReadValueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("returns nil for non-existent key", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		got, err := db.Read("test")

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
	t.Run("reads empty value", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		err = db.Write("empty-key", []byte{})

		// Assert
		assert.NoError(t, err)
		got, err := db.Read("empty-key")
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestDatabase_QueryValues(t *testing.T) {
	t.Run("panics when database not opened", func(t *testing.T) {
		// Arrange
		db := NewDatabase()
		// Act&Assert
		assert.Panics(t, func() { _, _ = db.Query("123") })
	})
}

func TestDatabase_QueryValuesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("return error when empty prefix", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		got, err := db.Query("")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestDatabase_DeleteValue(t *testing.T) {
	t.Run("panics when database not opened", func(t *testing.T) {
		// Arrange
		db := NewDatabase()
		// Act&Assert
		assert.Panics(t, func() { _ = db.Delete("123") })
	})
}

func TestDatabase_DeleteValueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("return nil for non-existent key", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		err = db.Delete("non-existent")

		// Assert
		assert.NoError(t, err)
	})
	t.Run("deletes existing value", func(t *testing.T) {
		// Arrange
		tf, cleanup := testTempFolder(t)
		defer cleanup()
		db := Database{BasePath: tf}
		defer closeTestDB(t, &db)
		err := db.Open("123")
		if err != nil {
			t.Fatal(err)
		}

		err = db.Write("key1", []byte("value1"))
		if err != nil {
			t.Fatal(err)
		}

		// Act
		err = db.Delete("key1")

		// Assert
		assert.NoError(t, err)
		value, err := db.Read("key1")
		assert.NoError(t, err)
		assert.Nil(t, value)
	})
}

func testTempFolder(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "signal-chat-test-*")
	if err != nil {
		t.Fatal(err)
	}
	return tempDir, func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	}
}

func closeTestDB(t *testing.T, db *Database) {
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
