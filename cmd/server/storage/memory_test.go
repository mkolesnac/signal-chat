package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Struct for testing
type TestItem struct {
	PartitionKey string
	SortKey      string
	Value        string
	SenderID     string
}

func (i *TestItem) GetPartitionKey() string {
	return i.PartitionKey
}

func (i *TestItem) GetSortKey() string {
	return i.SortKey
}

// Test for WriteItem and GetItem
func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	if storage == nil || storage.table == nil {
		t.Fatal("Expected non-nil MemoryStorage instance")
	}
}

// Test for GetItem with type mismatch
func TestMemoryStorage_GetItem(t *testing.T) {
	t.Run("when item exists", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		item := TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"}
		err := storage.WriteItem(&item)
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var result TestItem
		err = storage.GetItem("pk1", "sk1", &result)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, item, result)
	})
	t.Run("when item doesn't exist", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var result TestItem
		err := storage.GetItem("pk1", "sk1", &result)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("when outPtr is nil pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act + Assert
		var result *TestItem
		assert.Panics(t, func() { _ = storage.GetItem("pk1", "sk1", result) })
	})
	t.Run("when outPtr pointer has invalid type", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		item := TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"}
		err := storage.WriteItem(&item)
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act + Assert
		var wrongType int
		assert.Panics(t, func() { _ = storage.GetItem("pk1", "sk1", &wrongType) })
	})

}

// Test for QueryItems
func TestMemoryStorage_QueryItems(t *testing.T) {
	t.Run("success with QUERY_BEGINS_WITH query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "xx#1", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk#1", "sk", QUERY_BEGINS_WITH, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 2)
		var sortKeys []string
		for _, item := range results {
			sortKeys = append(sortKeys, item.SortKey)
		}
		assert.Equal(t, sortKeys[0], "sk#1")
		assert.Equal(t, sortKeys[1], "sk#2")
	})
	t.Run("success with QUERY_GREATER_THAN query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "xx#3", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk#1", "sk#0", QUERY_GREATER_THAN, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 2)
		assert.Equal(t, results[0].SortKey, "sk#1")
		assert.Equal(t, results[1].SortKey, "sk#2")
	})
	t.Run("success with QUERY_LOWER_THAN query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2"})
		err = storage.WriteItem(&TestItem{PartitionKey: "pk#1", SortKey: "xx#3", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk#1", "sk#2", QUERY_LOWER_THAN, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 1)
		assert.Equal(t, results[0].SortKey, "sk#1")
	})
	t.Run("panics when outSlicePtr not valid pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results []string
		assert.Panics(t, func() { _ = storage.QueryItems("pk#1", "sk#1", QUERY_BEGINS_WITH, results) })
	})
	t.Run("panics when outSlicePtr not slice pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results int
		assert.Panics(t, func() { _ = storage.QueryItems("pk#1", "sk#1", QUERY_BEGINS_WITH, &results) })
	})
	t.Run("panics when invalid query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results []TestItem
		assert.Panics(t, func() { _ = storage.QueryItems("pk#1", "sk#1", "invalid condition", &results) })
	})
}

func TestMemoryStorage_QueryItemsBySenderID(t *testing.T) {
	t.Run("success with QUERY_BEGINS_WITH query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		testItem1 := TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1", SenderID: "123"}
		testItem2 := TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test1", SenderID: "123"}
		storage.bySenderID["123:sk#1"] = &testItem1
		storage.bySenderID["123:sk#2"] = &testItem2

		// Act
		var results []TestItem
		err := storage.QueryItemsBySenderID("123", "sk", QUERY_BEGINS_WITH, &results)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		var sortKeys []string
		for _, item := range results {
			sortKeys = append(sortKeys, item.SortKey)
		}
		assert.Equal(t, sortKeys[0], "sk#1")
		assert.Equal(t, sortKeys[1], "sk#2")
	})
}

// Test for DeleteItem
func TestMemoryStorage_DeleteItem(t *testing.T) {
	t.Run("when item with SenderID field", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		testItem := TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test1", SenderID: "123"}
		storage.table["pk#1:sk#1"] = &testItem
		storage.bySenderID["123:sk#1"] = &testItem

		// Act
		err := storage.DeleteItem("pk#1", "sk#1")

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, storage.table)
		assert.Empty(t, storage.bySenderID)
	})
}

// Test for DeleteItem
func TestMemoryStorage_WriteItem(t *testing.T) {
	t.Run("when item with SenderID field", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		testItem := TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test1", SenderID: "123"}

		// Act
		err := storage.WriteItem(&testItem)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, storage.table)
		assert.NotEmpty(t, storage.bySenderID)
	})
}

// Test for BatchWriteItems
func TestMemoryStorage_BatchWriteItems(t *testing.T) {
	t.Run("writes all items on success", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		items := []PrimaryKeyProvider{
			&TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1"},
			&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2"},
		}

		// Act
		err := storage.BatchWriteItems(items)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, storage.table, 2)
	})
	t.Run("when some items contain SenderID", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		items := []PrimaryKeyProvider{
			&TestItem{PartitionKey: "pk#1", SortKey: "sk#1", Value: "test1"},
			&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2", SenderID: "123"},
		}

		// Act
		err := storage.BatchWriteItems(items)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, storage.table, 2)
		assert.Len(t, storage.bySenderID, 1)
	})
	t.Run("panics if one of the items is not pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		items := []PrimaryKeyProvider{
			nil,
			&TestItem{PartitionKey: "pk#1", SortKey: "sk#2", Value: "test2"},
		}

		// Act + Assert
		assert.Panics(t, func() { _ = storage.BatchWriteItems(items) })
	})
}
