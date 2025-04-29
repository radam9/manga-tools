package mangadex

import (
	"fmt"
	"github.com/radam9/manga-tools/internal/model"
	"io"
	"os"
)

func WritePagesToTempFiles(tempDir string, pages []model.Page) ([]model.Page, error) {
	var err error
	for i := range len(pages) {
		pages[i].Path, err = WritePageToTempFile(tempDir, pages[i])
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func WritePageToTempFile(tempDir string, page model.Page) (string, error) {
	fileName := fmt.Sprintf("%04d-*", page.Number)
	tempFile, err := os.CreateTemp(tempDir, fileName)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tempFile.Close()

	data, err := io.ReadAll(page.Data)
	if err != nil {
		return "", fmt.Errorf("reading page data: %w", err)
	}

	if _, err := tempFile.Write(data); err != nil {
		return "", fmt.Errorf("writing page to temp file: %w", err)
	}
	return tempFile.Name(), nil
}
