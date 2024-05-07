package utils

import (
	"fmt"
	"strings"
)

func StringJoin[T any](slice []T, sep string) string {
	if len(slice) == 0 {
		return ""
	}

	if len(slice) == 1 {
		return fmt.Sprint(slice[0])
	}

	var b strings.Builder
	b.WriteString(fmt.Sprint(slice[0]))
	for _, s := range slice[1:] {
		b.WriteString(sep)
		b.WriteString(fmt.Sprint(s))
	}

	return b.String()
}

func Map[T any, U any](slice []T, f func(T) U) []U {
	if len(slice) == 0 {
		return nil
	}

	result := make([]U, len(slice))
	for i, s := range slice {
		result[i] = f(s)
	}

	return result
}
