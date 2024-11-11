package storage

import (
	"reflect"
)

func panicIfNotPointer(outPtr any) {
	// Ensure that `ptr` is a non-nil pointer
	outValue := reflect.ValueOf(outPtr)
	if outValue.Kind() != reflect.Ptr || outValue.IsNil() {
		panic("outPtr parameter must be a non-nil pointer")
	}
}

func panicIfNotSlicePointer(outSlicePtr any) {
	// Ensure that `ptr` is a non-nil pointer
	outValue := reflect.ValueOf(outSlicePtr)
	if outValue.Kind() != reflect.Ptr || outValue.IsNil() {
		panic("outSlicePtr parameter must be a non-nil pointer")
	}

	// Ensure that the underlying value is a slice
	outElem := outValue.Elem()
	if outElem.Kind() != reflect.Slice {
		panic("outSlicePtr parameter must be a pointer to a slice")
	}
}

func panicIfInvalidQueryCondition(condition QueryCondition) {
	if condition != QUERY_BEGINS_WITH && condition != QUERY_GREATER_THAN && condition != QUERY_LOWER_THAN {
		panic("invalid query condition")
	}
}
