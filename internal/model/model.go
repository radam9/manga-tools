package model

import (
	"github.com/radam9/manga-tools/internal/ranges"
	"io"
	"slices"
)

type Chapter struct {
	ID         string
	Title      string
	Number     float64
	Volume     int
	PagesCount int
	Pages      []Page
	Language   string
}

type Page struct {
	Number int
	URL    string
	Data   io.Reader
	Path   FilePath
}

type FilePath = string

func SortChaptersByNumber(chapters []Chapter) {
	slices.SortStableFunc(chapters, func(a, b Chapter) int {
		if a.Number < b.Number {
			return -1
		} else if a.Number > b.Number {
			return 1
		}
		return 0
	})
}

func FilterChapters(chapters []Chapter, ranges []ranges.Range) []Chapter {
	var result []Chapter
	for _, chapter := range chapters {
		for _, rng := range ranges {
			if chapter.Number >= rng.Start && chapter.Number <= rng.End {
				result = append(result, chapter)
			}
		}
	}
	return result
}

func SortPagesByNumber(pages []Page) {
	slices.SortStableFunc(pages, func(a, b Page) int {
		if a.Number < b.Number {
			return -1
		} else if a.Number > b.Number {
			return 1
		}
		return 0
	})
}

func GetSliceOfPagePathsFromPages(pages []Page) []FilePath {
	var result []FilePath
	for _, page := range pages {
		result = append(result, page.Path)
	}
	return result
}
