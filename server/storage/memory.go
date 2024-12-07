package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	resources []Resource
	mu        sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		resources: make([]Resource, 0),
	}
}

func (s *MemoryStore) GetItem(pk, sk string) (Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, res := range s.resources {
		if res.PartitionKey == pk && res.SortKey == sk {
			return res, nil
		}
	}

	return Resource{}, ErrNotFound
}

func (s *MemoryStore) QueryItems(pk, sk string, queryCondition QueryCondition) ([]Resource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a function that will evaluate sort key matches based on the specified query condition
	skPrefix := strings.Split(sk, "#")[0]
	type FilterFunc = func(sk string) bool
	var skFilter FilterFunc
	if queryCondition == QueryBeginsWith {
		skFilter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix)
		}
	} else if queryCondition == QueryGreaterThan {
		skFilter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key > sk
		}
	} else if queryCondition == QueryLowerThan {
		skFilter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key < sk
		}
	}

	// Filter out resources based of the partition and sort keys
	var selected []Resource
	for _, res := range s.resources {
		if res.PartitionKey == pk && skFilter(res.SortKey) == true {
			selected = append(selected, res)
		}
	}

	// Sort the selected in ascending order by sort keys
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].SortKey < selected[j].SortKey
	})

	return selected, nil
}

func (s *MemoryStore) DeleteItem(pk, sk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.findResourceIndex(pk, sk)
	if index == -1 {
		// Item doesn't exist, no need to delete anything
		return nil
	}

	s.resources = append(s.resources[:index], s.resources[index+1:]...)
	return nil
}

func (s *MemoryStore) UpdateItem(pk, sk string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.findResourceIndex(pk, sk)
	if index == -1 {
		// Item doesn't exist, return error
		return ErrNotFound
	}

	item := s.resources[index]
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
	s.resources[index] = item
	return nil
}

func (s *MemoryStore) WriteItem(resource Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resources = append(s.resources, resource)
	return nil
}

func (s *MemoryStore) BatchWriteItems(resources []Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resources = append(s.resources, resources...)
	return nil
}

func (s *MemoryStore) findResourceIndex(pk, sk string) int {
	index := -1
	for i, res := range s.resources {
		if res.PartitionKey == pk && res.SortKey == sk {
			index = i
		}
	}

	return index
}
