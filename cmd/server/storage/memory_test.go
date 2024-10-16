package storage

import (
	"testing"
)

// Struct for testing
type TestItem struct {
	Name  string
	Value int
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
	item := TestItem{Name: "test", Value: 1}
	err := storage.WriteItem("pk1", "sk1", item)
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

// Test GetItem with invalid pointer (panic expected)
func TestGetItem_InvalidPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic for nil pointer")
		}
	}()

	// Act
	var result string
	_ = storage.GetItem("pk1", "sk1", result)
}

// Test for GetItem with type mismatch
func TestGetItem_TypeMismatch(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	err := storage.WriteItem("pk1", "sk1", "TestItem")
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
	err := storage.WriteItem("pk1", "profile1", TestItem{Name: "test1", Value: 1})
	err = storage.WriteItem("pk1", "profile2", TestItem{Name: "test2", Value: 2})
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	// Act
	var results []TestItem
	err = storage.QueryItems("pk1", "profile", &results)
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
	err := storage.WriteItem("pk1", "sk1", "Item1")
	if err != nil {
		t.Fatalf("WriteItem failed: %v", err)
	}

	// Act
	err = storage.DeleteItem("pk1", "sk1")
	if err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	// Assert
	var result string
	err = storage.GetItem("pk1", "sk1", &result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "" {
		t.Fatalf("Expected no item after deletion, got %v", result)
	}
}

// Test for BatchWriteItems
func TestBatchWriteItems(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	items := []BatchWriteItem{
		{PartitionKey: "pk1", SortKey: "sk1", Value: TestItem{Name: "test1", Value: 1}},
		{PartitionKey: "pk1", SortKey: "sk2", Value: TestItem{Name: "test2", Value: 2}},
	}

	// Act
	err := storage.BatchWriteItems(items)
	if err != nil {
		t.Fatalf("BatchWriteItems failed: %v", err)
	}

	// Assert
	var result TestItem
	_ = storage.GetItem("pk1", "sk1", &result)
	if result != items[0].Value {
		t.Fatalf("Expected %v, got %v", items[0].Value, result)
	}
}

// Test for BatchWriteItems with pointer value (should return error)
func TestBatchWriteItems_WithPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	items := []BatchWriteItem{
		{PartitionKey: "pk1", SortKey: "sk1", Value: TestItem{Name: "test1", Value: 1}},
		{PartitionKey: "pk1", SortKey: "sk2", Value: &TestItem{Name: "test2", Value: 2}},
	}

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic when value field is a pointer")
		}
	}()

	// Act
	_ = storage.BatchWriteItems(items)
}

// Test for GetItem with invalid pointer
func TestWriteItem_WithPointer(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()

	defer func() {
		// Assert
		if r := recover(); r == nil {
			t.Fatal("Expected panic value arg is pointer")
		}
	}()

	// Act
	item := TestItem{Name: "Test", Value: 42}
	_ = storage.WriteItem("pk1", "sk1", &item)
}
