package format

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/radam9/manga-tools/internal/model"
	"io"
	"os"
	"path/filepath"
)

type CBR struct{}

func (c CBR) Save(filePath string, pages []model.FilePath) error {
	return saveAsCBArchive(filePath, pages)
}

func (c CBR) OutputPath(outputDir string, mangaTitle string, volume int, chapterTitle string, chapter float64) string {
	filePath := getOutputFilePath(outputDir, mangaTitle, volume, chapter, chapterTitle)
	return filePath + ".cbr"
}

type CBZ struct{}

func (c CBZ) Save(filePath string, pages []model.FilePath) error {
	return saveAsCBArchive(filePath, pages)
}

func (c CBZ) OutputPath(outputDir string, mangaTitle string, volume int, chapterTitle string, chapter float64) string {
	filePath := getOutputFilePath(outputDir, mangaTitle, volume, chapter, chapterTitle)
	return filePath + ".cbz"
}

func saveAsCBArchive(filePath string, pages []model.FilePath) error {
	if len(pages) == 0 {
		return errors.New("no files to pack")
	}
	buff, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer buff.Close()
	w := zip.NewWriter(buff)

	for i, page := range pages {
		if err := writeFileToArchive(w, i, page); err != nil {
			return err
		}
	}
	return w.Close()
}

func writeFileToArchive(w *zip.Writer, pageNum int, pagePath model.FilePath) error {
	f, err := w.Create(fmt.Sprintf("%04d.jpg", pageNum))
	if err != nil {
		return err
	}

	pageData, err := os.Open(pagePath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, pageData); err != nil {
		return err
	}
	return os.Remove(pagePath)
}

func ExtractArchive(tempDir, archiveDir, archiveName string) ([]model.FilePath, error) {
	archivePath := filepath.Join(archiveDir, archiveName)
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("opening archive %q: %w", archivePath, err)
	}
	defer r.Close()

	var imagePaths []string
	outputDir := filepath.Join(tempDir, archiveName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory %q: %w", outputDir, err)
	}

	for _, file := range r.File {
		dstFile, err := os.Create(filepath.Join(outputDir, file.Name))
		if err != nil {
			return nil, fmt.Errorf("opening temp file %q: %w", file.Name, err)
		}

		fileInArchive, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("opening archive file %q: %w", file.Name, err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return nil, fmt.Errorf("copying file %q: %w", file.Name, err)
		}

		imagePaths = append(imagePaths, dstFile.Name())
		dstFile.Close()
		fileInArchive.Close()
	}
	return imagePaths, nil
}
