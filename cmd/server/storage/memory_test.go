package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test for WriteItem and GetItem
func TestNewMemoryStore(t *testing.T) {
	storage := NewMemoryStore()
	if storage == nil || storage.resources == nil {
		t.Fatal("Expected non-nil MemoryStore instance")
	}
}

// Test for GetItem with type mismatch
func TestMemoryStore_GetItem(t *testing.T) {
	t.Run("returns error when item doesn't exist", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()

		// Act
		item, err := memStore.GetItem("pk1", "sk1")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, "", item.PartitionKey)
	})
	t.Run("test success when item exists", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		item, err := memStore.GetItem("acc#123", "conv#abs")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "acc#123", item.PartitionKey)
		assert.Equal(t, "conv#abs", item.SortKey)
	})
}

// Test for QueryItems
func TestMemoryStore_QueryItems(t *testing.T) {
	t.Run("test success with QUERY_BEGINS_WITH query", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		items, err := memStore.QueryItems("conv#abs", "", QueryBeginsWith)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, items, 4)
		for _, item := range items {
			assert.Equal(t, "conv#abs", item.PartitionKey)
			// Check whether the sort key if one the specified sort keys
			assert.Contains(t, []string{"conv#abs", "acc#123", "msg#1", "msg#2"}, item.SortKey)
		}
	})
	t.Run("success with QUERY_GREATER_THAN query condition", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		items, err := memStore.QueryItems("conv#abs", "msg#1", QueryGreaterThan)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "msg#2", items[0].SortKey)
	})
	t.Run("success with QUERY_LOWER_THAN query condition", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		items, err := memStore.QueryItems("conv#abs", "msg#2", QueryLowerThan)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "msg#1", items[0].SortKey)
	})
}

// Test for DeleteItem
func TestMemoryStore_DeleteItem(t *testing.T) {
	t.Run("returns no error when item doesn't exist", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()

		// Act
		err := memStore.DeleteItem("pk#1", "sk#1")

		// Assert
		assert.NoError(t, err)
	})
	t.Run("removes item from storage when it exists", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		err := memStore.DeleteItem("acc#123", "acc#123")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, memStore.resources, len(testResources)-1)
		// Assert that the account resource was deleted
		assert.NotContains(t, memStore.resources, testResources[0], "item was not removed from the internal storage")
	})
}

func TestMemoryStore_UpdateItem(t *testing.T) {
	t.Run("returns error when item doesn't exist", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()

		// Act
		err := memStore.UpdateItem("pk#1", "sk#1", map[string]interface{}{"Name": "test"})

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("returns error when updating not-existing field", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		err := memStore.UpdateItem("acc#123", "acc#123", map[string]interface{}{"NotExisting": "test"})

		// Assert
		assert.Error(t, err)
		assert.Equal(t, testResources, memStore.resources) // assert that resource weren't changes
	})
	t.Run("returns error when updated value type doesn't match field type", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		err := memStore.UpdateItem("acc#123", "acc#123", map[string]interface{}{"Name": 1})

		// Assert
		assert.Error(t, err)
		assert.Equal(t, testResources, memStore.resources) // assert that resource weren't changes
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		memStore := NewMemoryStore()
		memStore.resources = append(memStore.resources, testResources...)

		// Act
		sid := "test"
		err := memStore.UpdateItem("acc#123", "acc#123", map[string]interface{}{"SignedPreKeyID": &sid})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test", *memStore.resources[0].SignedPreKeyID) // assert the account resource was updated
	})
}

// Test for DeleteItem
func TestMemoryStore_WriteItem(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStore()

		// Act
		err := storage.WriteItem(testResources[0])

		// Assert
		assert.NoError(t, err)
		assert.Len(t, storage.resources, 1)
		assert.Equal(t, testResources[0], storage.resources[0])
	})
}

// Test for BatchWriteItems
func TestMemoryStore_BatchWriteItems(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		// Arrange
		storage := NewMemoryStore()

		// Act
		err := storage.BatchWriteItems(testResources)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, storage.resources, len(testResources))
		assert.Equal(t, testResources, storage.resources)
	})
}
