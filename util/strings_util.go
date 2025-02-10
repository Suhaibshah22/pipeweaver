package util

import (
	"slices"
	"strings"
)

func FilterByPrefix(words []string, prefix string) []string {
	return slices.Clip(slices.DeleteFunc(words, func(word string) bool {
		return !strings.HasPrefix(word, prefix) // Inverting logic 😑
	}))
}
