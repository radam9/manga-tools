package internal

import (
	"os"
	"slices"
)

func SortDirEntry(items []os.DirEntry) []os.DirEntry {
	slices.SortStableFunc(items, func(a, b os.DirEntry) int {
		// dir is always before a file
		if a.IsDir() && !b.IsDir() {
			return -1
		} else if !a.IsDir() && b.IsDir() {
			return 1
		}

		if a.Name() < b.Name() {
			return -1
		} else if a.Name() > b.Name() {
			return 1
		}
		return 0

	})
	return items
}
