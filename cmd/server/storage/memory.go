package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type MemoryStorage struct {
	table      map[string]any
	bySenderID map[string]any
	mu         sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		table:      make(map[string]any),
		bySenderID: make(map[string]any),
	}
}

func (s *MemoryStorage) GetItem(pk, sk string, outPtr any) error {
	panicIfNotPointer(outPtr)

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	item, ok := s.table[key]
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
	return s.runMemoryQuery(pk, sk, queryCondition, outSlicePtr, s.table)
}

func (s *MemoryStorage) QueryItemsBySenderID(senderID, sk string, queryCondition QueryCondition, outSlicePtr any) error {
	return s.runMemoryQuery(senderID, sk, queryCondition, outSlicePtr, s.bySenderID)
}

func (s *MemoryStorage) DeleteItem(pk, sk string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(pk, sk)
	item, ok := s.table[key]
	if !ok {
		return nil
	}

	delete(s.table, key)
	if senderID, ok := getSenderID(item); ok {
		indexKey := getPrimaryKey(senderID, sk)
		delete(s.bySenderID, indexKey)
	}

	return nil
}

func (s *MemoryStorage) WriteItem(item PrimaryKeyProvider) error {
	panicIfNotPointer(item)

	s.mu.Lock()
	defer s.mu.Unlock()

	key := getPrimaryKey(item.GetPartitionKey(), item.GetSortKey())
	s.table[key] = item

	if senderID, ok := getSenderID(item); ok {
		indexKey := getPrimaryKey(senderID, item.GetSortKey())
		s.bySenderID[indexKey] = item
	}
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
		s.table[key] = item

		if senderID, ok := getSenderID(item); ok {
			indexKey := getPrimaryKey(senderID, item.GetSortKey())
			s.bySenderID[indexKey] = item
		}
	}

	return nil
}

func getPrimaryKey(pk, sk string) string {
	return fmt.Sprintf("%s:%s", pk, sk)
}

func getSenderID(input interface{}) (string, bool) {
	val := reflect.ValueOf(input)

	// Check if the input is a struct or a pointer to a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference the pointer
	}
	if val.Kind() != reflect.Struct {
		return "", false // Not a struct, return false
	}

	// Check if the struct has a field named "SenderID"
	field := val.FieldByName("SenderID")
	if !field.IsValid() {
		return "", false // Field does not exist
	}

	// Check if the field is a string
	if field.Kind() == reflect.String {
		value := field.String()
		if value == "" {
			return "", false
		}

		return value, true
	}

	return "", false // Field exists but is not a string
}

func (s *MemoryStorage) runMemoryQuery(pk, sk string, queryCondition QueryCondition, outSlicePtr any, table map[string]any) error {
	panicIfNotSlicePointer(outSlicePtr)
	panicIfInvalidQueryCondition(queryCondition)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a function that will evaluate sort key matches based on the specified query condition
	skPrefix := fmt.Sprintf("%v#", strings.Split(sk, "#")[0])
	type FilterFunc = func(sk string) bool
	var filter FilterFunc
	if queryCondition == QUERY_BEGINS_WITH {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix)
		}
	} else if queryCondition == QUERY_GREATER_THAN {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key > sk
		}
	} else if queryCondition == QUERY_LOWER_THAN {
		filter = func(key string) bool {
			return strings.HasPrefix(key, skPrefix) && key < sk
		}
	}

	// Filter out items based of the partition and sort keys
	var items []interface{}
	for k, v := range table {
		parts := strings.Split(k, ":")
		if pk == parts[0] && filter(parts[1]) {
			items = append(items, v)
		}
	}

	// Sort the items in ascending order by sort keys
	sort.Slice(items, func(i, j int) bool {
		itemI := items[i].(PrimaryKeyProvider)
		itemJ := items[j].(PrimaryKeyProvider)

		return itemI.GetSortKey() < itemJ.GetSortKey()
	})

	// Convert outSlicePtr (a pointer to a items) into a reflection interface
	outSlicePtrValue := reflect.ValueOf(outSlicePtr)
	// Dereference the pointer to access the underlying items value
	outSliceValue := outSlicePtrValue.Elem()
	// Add items to the output slice using reflection
	for _, v := range items {
		item := reflect.ValueOf(v).Elem()
		outSliceValue.Set(reflect.Append(outSliceValue, item))
	}

	return nil
}
