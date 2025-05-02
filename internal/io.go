package internal

import (
	"github.com/maruel/natural"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func SortDirEntry(items []os.DirEntry) []os.DirEntry {
	slices.SortStableFunc(items, func(a, b os.DirEntry) int {
		// dir is always before a file
		if a.IsDir() && !b.IsDir() {
			return -1
		} else if !a.IsDir() && b.IsDir() {
			return 1
		}

		aName := strings.TrimSuffix(a.Name(), filepath.Ext(a.Name()))
		bName := strings.TrimSuffix(b.Name(), filepath.Ext(b.Name()))

		if natural.Less(aName, bName) {
			return -1
		} else if natural.Less(bName, aName) {
			return 1
		}
		return 0
	})
	return items
}
