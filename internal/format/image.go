package format

import (
	"errors"
	"fmt"
	"github.com/radam9/manga-tools/internal/model"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

type Image struct{}

func (i Image) Save(outputPath string, pages []model.FilePath) error {
	for index, page := range pages {
		if err := os.MkdirAll(outputPath, 0755); err != nil {
			return fmt.Errorf("save pages as images: mkdirall: %w", err)
		}

		outputFilename := filepath.Join(outputPath, fmt.Sprintf("%04d.png", index))
		err := os.Rename(page, outputFilename)
		if err == nil {
			continue
		}
		if !errors.Is(err, syscall.EXDEV) {
			return fmt.Errorf("save pages as images: move: %w", err)
		}

		if err := i.crossDeviceCopy(outputFilename, pages[index]); err != nil {
			return fmt.Errorf("save pages as images: copy: %w", err)
		}
	}
	return nil
}

func (i Image) OutputPath(outputDir string, mangaTitle string, volume int, chapterTitle string, chapter float64) string {
	return getOutputFilePath(outputDir, mangaTitle, volume, chapter, chapterTitle)
}

func (i Image) crossDeviceCopy(outputPath string, page model.FilePath) error {
	src, err := os.Open(page)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return os.Remove(page)
}
