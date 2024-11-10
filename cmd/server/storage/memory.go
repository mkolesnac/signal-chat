package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type MemoryStorage struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]any),
	}
}

func (s *MemoryStorage) GetItem(pk, sk string, outPtr any) error {
	panicIfNotPointer(outPtr)

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	item, ok := s.data[key]
	if !ok {
		return ErrNotFound
	}
	// Set the item of `outPtr` to the item retrieved from the map
	outValue := reflect.ValueOf(outPtr)
	itemValue := reflect.ValueOf(item)
	if !itemValue.Type().AssignableTo(outValue.Type()) {
		panic("type mismatch: cannot assign item to outPtr parameter")
	}

	outValue.Elem().Set(itemValue.Elem())

	return nil
}

func (s *MemoryStorage) QueryItems(pk, sk string, queryCondition QueryCondition, outSlicePtr any) error {
	panicIfNotSlicePointer(outSlicePtr)
	panicIfInvalidQueryCondition(queryCondition)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert outSlicePtr (a pointer to a slice) into a reflection interface
	outValue := reflect.ValueOf(outSlicePtr)
	// Dereference the pointer to access the underlying slice value
	sliceValue := outValue.Elem()

	// Create a function that will evaluate sort key matches based on the specified query condition
	skPrefix := fmt.Sprintf("%v#", strings.Split(sk, "#")[0])
	type FilterFunc = func(sk string) bool
	var filter FilterFunc
	if queryCondition == BEGINS_WITH {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix)
		}
	} else if queryCondition == GREATER_THAN {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key > sk
		}
	} else if queryCondition == LOWER_THAN {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key < sk
		}
	}

	for k, v := range s.data {
		parts := strings.Split(k, ":")
		if pk == parts[0] && filter(parts[1]) {
			item := reflect.ValueOf(v).Elem()
			sliceValue.Set(reflect.Append(sliceValue, item))
		}
	}

	// Sort the slice in ascending order by sort keys
	sort.Slice(sliceValue.Interface(), func(i, j int) bool {
		itemI := sliceValue.Index(i).Interface().(PrimaryKeyProvider)
		itemJ := sliceValue.Index(j).Interface().(PrimaryKeyProvider)

		return itemI.GetSortKey() < itemJ.GetSortKey()
	})

	return nil
}

func (s *MemoryStorage) DeleteItem(pk, sk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	delete(s.data, key)
	return nil
}

func (s *MemoryStorage) WriteItem(item PrimaryKeyProvider) error {
	panicIfNotPointer(item)

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(item.GetPartitionKey(), item.GetSortKey())
	s.data[key] = item
	return nil
}

func (s *MemoryStorage) BatchWriteItems(items []PrimaryKeyProvider) error {
	for _, item := range items {
		panicIfNotPointer(item)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		key := getPrimaryKey(item.GetPartitionKey(), item.GetSortKey())
		s.data[key] = item
	}

	return nil
}

func getPrimaryKey(pk, sk string) string {
	return fmt.Sprintf("%s:%s", pk, sk)
}
