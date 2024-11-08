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
}

func (i TestItem) GetPartitionKey() string {
	return i.PartitionKey
}

func (i TestItem) GetSortKey() string {
	return i.SortKey
}

// Test for WriteItem and GetItem
func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	if storage == nil || storage.data == nil {
		t.Fatal("Expected non-nil MemoryStorage instance")
	}
}

func TestWriteAndGetItem(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	item := TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"}
	err := storage.WriteItem(item)
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	// Act
	var result TestItem
	err = storage.GetItem("pk1", "sk1", &result)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}

	// Assert
	if result != item {
		t.Fatalf("Expected %v, got %v", item, result)
	}
}

func TestGetItem_ItemDoesntExist(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	// Act
	var result TestItem
	err := storage.GetItem("pk1", "sk1", &result)
	// Assert
	if err == nil {
		t.Fatalf("Expected GetItem to return error but got %v", err)
	}
}

// Test GetItem with invalid pointer (panic expected)
func TestGetItem_InvalidPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for pointer of invalid type")
		}
	}()

	// Act
	var result string
	_ = storage.GetItem("pk1", "sk1", result)
}

// Test GetItem with nil pointer (panic expected)
func TestGetItem_NilPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for nil pointer")
		}
	}()

	// Act
	var result *TestItem
	_ = storage.GetItem("pk1", "sk1", result)
}

// Test for GetItem with type mismatch
func TestGetItem_TypeMismatch(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	item := TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"}
	err := storage.WriteItem(item)
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for pointer of invalid type")
		}
	}()

	// Act
	var wrongType int
	err = storage.GetItem("pk1", "sk1", &wrongType)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
}

// Test for QueryItems
func TestQueryItems(t *testing.T) {
	t.Run("success with BEGINS_WITH query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "xx", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk1", "sk", BEGINS_WITH, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 2)
		var sortKeys []string
		for _, item := range results {
			sortKeys = append(sortKeys, item.SortKey)
		}
		assert.Equal(t, sortKeys[0], "sk1")
		assert.Equal(t, sortKeys[1], "sk2")
	})
	t.Run("success with GREATER_THAN query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk3", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk1", "sk2", GREATER_THAN, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 1)
		assert.Equal(t, results[0].SortKey, "sk3")
	})
	t.Run("success with LOWER_THAN query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()
		err := storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"})
		err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk3", Value: "test3"})
		if err != nil {
			t.Fatalf("WriteItem failed: %v", err)
		}

		// Act
		var results []TestItem
		err = storage.QueryItems("pk1", "sk2", LOWER_THAN, &results)
		if err != nil {
			t.Fatalf("QueryItems failed: %v", err)
		}

		// Assert
		assert.Len(t, results, 1)
		assert.Equal(t, results[0].SortKey, "sk1")
	})
	t.Run("panics when outSlicePtr not valid pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results []string
		assert.Panics(t, func() { _ = storage.QueryItems("pk1", "profile", BEGINS_WITH, results) })
	})
	t.Run("panics when outSlicePtr not slice pointer", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results int
		assert.Panics(t, func() { _ = storage.QueryItems("pk1", "profile", BEGINS_WITH, &results) })
	})
	t.Run("panics when invalid query condition", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStorage()

		// Act
		var results []TestItem
		assert.Panics(t, func() { _ = storage.QueryItems("pk1", "sk", "invalid condition", &results) })
	})
}

// Test for DeleteItem
func TestDeleteItem(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	err := storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"})
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	// Act
	err = storage.DeleteItem("pk1", "sk1")
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	// Assert
	assert.Empty(t, storage.data)
}

// Test for BatchWriteItems
func TestBatchWriteItems(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	items := []PrimaryKeyProvider{
		TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"},
		TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"},
	}

	// Act
	err := storage.BatchWriteItems(items)
	if err != nil {
		t.Fatalf("BatchWriteItems failed: %v", err)
	}

	// Assert
	assert.Len(t, storage.data, 2)
}
