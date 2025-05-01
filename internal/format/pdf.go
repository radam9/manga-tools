package format

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	model2 "github.com/radam9/manga-tools/internal/model"
)

type PDF struct{}

func (p PDF) Save(filePath string, pages []model2.FilePath) error {
	if len(pages) == 0 {
		return nil
	}
	imp, conf := DefaultPDFConfig()
	err := api.ImportImagesFile(pages, filePath, imp, conf)
	if err != nil {
		return fmt.Errorf("creating pdf from images: %w", err)
	}

	err = api.OptimizeFile(filePath, filePath, conf)
	if err != nil {
		return fmt.Errorf("optimizing pdf file: %w", err)
	}
	return nil
}

func (p PDF) OutputPath(outputDir string, mangaTitle string, volume int, chapterTitle string, chapter float64) string {
	outputPath := getOutputFilePath(outputDir, mangaTitle, volume, chapter, chapterTitle)
	return outputPath + ".pdf"
}

func DefaultPDFConfig() (*pdfcpu.Import, *model.Configuration) {
	imp := pdfcpu.DefaultImportConfig()
	imp.DPI = 300
	imp.Scale = 1
	conf := model.NewDefaultConfiguration()
	conf.OptimizeDuplicateContentStreams = true
	return imp, conf
}
