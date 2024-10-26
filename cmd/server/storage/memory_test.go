package storage

import (
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
	// Arrange
	storage := NewMemoryStorage()
	err := storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"})
	err = storage.WriteItem(TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"})
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	// Act
	var results []TestItem
	err = storage.QueryItems("pk1", "sk", &results)
	if err != nil {
		t.Fatalf("QueryItems failed: %v", err)
	}

	// Assert
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// Test case where outSlicePtr is not a pointer
func TestQueryItems_NotPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for non-pointer slice")
		}
	}()

	// Act
	var results []string
	_ = storage.QueryItems("pk1", "profile", results)
}

// Test case where outSlicePtr is not a slice
func TestQueryItems_NotSlice(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for non-slice pointer")
		}
	}()

	// Act
	var results int
	_ = storage.QueryItems("pk1", "sk", &results)
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
	var result TestItem
	err = storage.GetItem("pk1", "sk1", &result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.Value != "" {
		t.Fatalf("Expected no item after deletion, got %v", result)
	}
}

// Test for BatchWriteItems
func TestBatchWriteItems(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	items := []WriteableItem{
		TestItem{PartitionKey: "pk1", SortKey: "sk1", Value: "test1"},
		TestItem{PartitionKey: "pk1", SortKey: "sk2", Value: "test2"},
	}

	// Act
	err := storage.BatchWriteItems(items)
	if err != nil {
		t.Fatalf("BatchWriteItems failed: %v", err)
	}

	// Assert
	var result TestItem
	_ = storage.GetItem("pk1", "sk1", &result)
	if result != items[0] {
		t.Fatalf("Expected %v, got %v", items[0], result)
	}
}
