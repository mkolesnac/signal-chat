package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	items map[string]Resource
	mu    sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]Resource),
	}
}

func (s *MemoryStore) GetItem(pk, sk string) (Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	item, ok := s.items[key]
	if !ok {
		return Resource{}, ErrNotFound
	}

	return item, nil
}

func (s *MemoryStore) QueryItems(pk, sk string, queryCondition QueryCondition) ([]Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a function that will evaluate sort key matches based on the specified query condition
	skPrefix := fmt.Sprintf("%v#", strings.Split(sk, "#")[0])
	type FilterFunc = func(sk string) bool
	var filter FilterFunc
	if queryCondition == QueryBeginsWith {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix)
		}
	} else if queryCondition == QueryGreaterThan {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key > sk
		}
	} else if queryCondition == QueryLowerThan {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key < sk
		}
	}

	// Filter out resources based of the partition and sort keys
	var resources []Resource
	for k, v := range s.items {
		parts := strings.Split(k, ":")
		if pk == parts[0] && filter(parts[1]) {
			resources = append(resources, v)
		}
	}

	// Sort the resources in ascending order by sort keys
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].SortKey < resources[j].SortKey
	})

	return resources, nil
}

func (s *MemoryStore) DeleteItem(pk, sk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	delete(s.items, key)

	return nil
}

func (s *MemoryStore) UpdateItem(pk, sk string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	item, ok := s.items[key]
	if !ok {
		return ErrNotFound
	}

	// Get the reflect.Value of the struct
	v := reflect.ValueOf(&item).Elem() // Get a pointer to the value to make it addressable

	// Iterate over the updates and set the fields
	for fieldName, value := range updates {
		field := v.FieldByName(fieldName)

		// Check if the field exists and is settable
		if !field.IsValid() {
			return fmt.Errorf("no such field: %s in target struct", fieldName)
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field: %s", fieldName)
		}

		// Check type compatibility
		fieldValue := reflect.ValueOf(value)
		if field.Type() != fieldValue.Type() {
			return fmt.Errorf("type mismatch for field: %s", fieldName)
		}

		// Update the field
		field.Set(fieldValue)
	}

	// Write the updated struct back into the map
	s.items[key] = item
	return nil
}

func (s *MemoryStore) WriteItem(resource Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(resource.PartitionKey, resource.SortKey)
	s.items[key] = resource

	return nil
}

func (s *MemoryStore) BatchWriteItems(resources []Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range resources {
		key := getPrimaryKey(r.PartitionKey, r.SortKey)
		s.items[key] = r
	}

	return nil
}

func getPrimaryKey(pk, sk string) string {
	return fmt.Sprintf("%s:%s", pk, sk)
}
