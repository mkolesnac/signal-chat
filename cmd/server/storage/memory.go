package storage

import (
	"errors"
	"fmt"
	"reflect"
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
		return errors.New("key not found")
	}
	// Set the item of `outPtr` to the item retrieved from the map
	outValue := reflect.ValueOf(outPtr)
	itemValue := reflect.ValueOf(item)
	if !itemValue.Type().AssignableTo(outValue.Elem().Type()) {
		panic("type mismatch: cannot assign item to outPtr parameter")
	}

	outValue.Elem().Set(itemValue)

	return nil
}

func (s *MemoryStorage) QueryItems(pk, skPrefix string, outSlicePtr any) error {
	panicIfNotSlicePointer(outSlicePtr)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert outSlicePtr (a pointer to a slice) into a reflection interface
	outValue := reflect.ValueOf(outSlicePtr)
	// Dereference the pointer to access the underlying slice value
	sliceValue := outValue.Elem()

	for k, v := range s.data {
		parts := strings.Split(k, ":")
		if pk == parts[0] && strings.HasPrefix(parts[1], skPrefix) {
			item := reflect.ValueOf(v)
			sliceValue.Set(reflect.Append(sliceValue, item))
		}
	}

	return nil
}

func (s *MemoryStorage) DeleteItem(pk, sk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	delete(s.data, key)
	return nil
}

func (s *MemoryStorage) WriteItem(item WriteableItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(item.GetPartitionKey(), item.GetSortKey())
	s.data[key] = item
	return nil
}

func (s *MemoryStorage) BatchWriteItems(items []WriteableItem) error {
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
