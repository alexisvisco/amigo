package utils

import "reflect"

// Ptr returns a pointer to the value passed as argument.
func Ptr[T any](t T) *T { return &t }

// NilOrValue If the value is a default value for the type, it returns nil else it returns the value.
func NilOrValue[T any](t T) *T {
	var zero T
	if reflect.DeepEqual(t, zero) {
		return nil
	}
	return &t
}
