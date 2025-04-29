package internal

import (
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestSortDirEntry(t *testing.T) {
	type dirEntry struct {
		name string
		dir  bool
	}

	filesystem := &fstest.MapFS{
		"root/third":        &fstest.MapFile{Mode: fs.ModeDir},
		"root/third/file3":  &fstest.MapFile{Mode: fs.ModeAppend},
		"root/third/file1":  &fstest.MapFile{Mode: fs.ModeAppend},
		"root/third/file2":  &fstest.MapFile{Mode: fs.ModeAppend},
		"root/first/file1":  &fstest.MapFile{Mode: fs.ModeAppend},
		"root/first":        &fstest.MapFile{Mode: fs.ModeDir},
		"root/first/file2":  &fstest.MapFile{Mode: fs.ModeAppend},
		"root/second":       &fstest.MapFile{Mode: fs.ModeDir},
		"root/second/file1": &fstest.MapFile{Mode: fs.ModeAppend},
		"root/file3":        &fstest.MapFile{Mode: fs.ModeAppend},
		"root/file2":        &fstest.MapFile{Mode: fs.ModeAppend},
		"root/file1":        &fstest.MapFile{Mode: fs.ModeAppend},
		"root":              &fstest.MapFile{Mode: fs.ModeDir},
		"root/zero":         &fstest.MapFile{Mode: fs.ModeDir},
		"root/afile0":       &fstest.MapFile{Mode: fs.ModeAppend},
	}

	children, err := filesystem.ReadDir("root")
	if err != nil {
		t.Fatal(err)
	}

	expected := []dirEntry{
		{name: "first", dir: true},
		{name: "second", dir: true},
		{name: "third", dir: true},
		{name: "zero", dir: true},
		{name: "afile0", dir: false},
		{name: "file1", dir: false},
		{name: "file2", dir: false},
		{name: "file3", dir: false},
	}

	got := SortDirEntry(children)

	for i := range 8 {
		expectedItem := expected[i]
		gotItem := got[i]
		fmt.Printf("expected: %q, got: %q\n", expectedItem.name, gotItem.Name())
		if expectedItem.name != gotItem.Name() {
			t.Errorf("expected: %s, got: %s", expectedItem.name, gotItem.Name())
		}
	}

}
