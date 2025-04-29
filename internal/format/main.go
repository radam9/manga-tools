package format

import (
	"fmt"
	"github.com/radam9/manga-tools/internal/model"
	"path/filepath"
	"strings"
)

func getOutputFilePath(outputDir string, mangaTitle string, volume int, chapter float64, chapterTitle string) string {
	var filename strings.Builder
	filename.WriteString(mangaTitle)

	if volume > 0 {
		filename.WriteString(fmt.Sprintf(" - volume %d", volume))
	}
	if chapter > 0 {
		filename.WriteString(fmt.Sprintf(" - chapter %06.1f", chapter))
	}
	if chapterTitle != "" {
		filename.WriteString(fmt.Sprintf(" - %s", chapterTitle))
	}
	return filepath.Join(outputDir, filename.String())
}

type Format interface {
	Save(outputPath string, pages []model.FilePath) error
	OutputPath(outputDir string, mangaTitle string, volume int, chapterTitle string, chapter float64) string
}

func SelectFormat(cbr, cbz, pdf bool) Format {
	if pdf {
		return PDF{}
	} else if cbr {
		return CBR{}
	} else if cbz {
		return CBZ{}
	}
	return Image{}
}
