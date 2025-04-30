package cmd

import (
	"fmt"
	"github.com/radam9/manga-tools/internal"
	"github.com/radam9/manga-tools/internal/format"
	"github.com/radam9/manga-tools/internal/model"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var imageFormats = []string{".jpg", ".jpeg", ".png"}
var archiveFormats = []string{".cbr", ".cbz"}

type convertOptions struct {
	dir     string
	files   []string
	archive bool
	image   bool
	bundle  bool
}

func NewConvertCommand() *cobra.Command {
	options := &convertOptions{}

	cmd := &cobra.Command{
		Use:   "convert [OPTIONS]",
		Short: "convert a set of images, cbr or cbz files to pdf",
		Long: `convert a set of images, cbr or cbz files to pdf.
By default the command will try to convert images to pdf, pass the appropriate flag to convert to a different format.

The items are converted to a pdf and bundled following the rules below:
- Each sub-directory of the given directory will be converted to a separate pdf.
- All children files of the given directory will be converted to a single pdf if they are images,
	If they are archives, each archive will be converted to separate pdf.

If the bundle flag is passed, the everything will bundled in a single pdf in the following order:
	- The items of each sub-directory of the provided directory will be ordered by path and 
		the files in the root of the dir will be appended at the end.`,
		Args: cobra.NoArgs,
		RunE: convertCommandRunFunction(options),
	}

	flags := cmd.Flags()

	const convertDirFlag = "dir"
	flags.StringVarP(&options.dir, convertDirFlag, "d", "", "path to the directory containing files to convert")
	err := cmd.MarkFlagRequired(convertDirFlag)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	const convertArchiveFlag = "archive"
	const convertImageFlag = "image"
	flags.BoolVar(&options.archive, convertArchiveFlag, false, "source files are archive format (cbr/cbz)")
	flags.BoolVar(&options.image, convertImageFlag, false, "source files are image format")
	cmd.MarkFlagsMutuallyExclusive(convertArchiveFlag, convertImageFlag)

	flags.BoolVarP(&options.bundle, "bundle", "b", false, "bundle all passed dir and files into a single pdf")
	return cmd
}

const inputTypeImage = "image"
const inputTypeArchive = "archive"

type OutputUnit struct {
	dir   string
	files []model.FilePath
}

func (o OutputUnit) getImages(tempDir string, inputFormat string) ([]string, error) {
	var images []string
	if inputFormat == inputTypeImage {
		for _, image := range o.files {
			images = append(images, image)
		}
		return images, nil
	}

	// archive mode
	for _, archive := range o.files {
		imagePaths, err := format.ExtractArchive(tempDir, o.dir, filepath.Base(archive))
		if err != nil {
			return nil, fmt.Errorf("extracting archive %q: %w", archive, err)
		}
		images = append(images, imagePaths...)
	}
	return images, nil
}

func (o *OutputUnit) appendFile(newFile model.FilePath) {
	currentFiles := o.files
	currentFiles = append(currentFiles, newFile)
	o.files = currentFiles
}

func convertCommandRunFunction(options *convertOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		allowedFormats := imageFormats
		inputFormat := inputTypeImage
		if options.archive {
			allowedFormats = archiveFormats
			inputFormat = inputTypeArchive
		}

		pdf := format.PDF{}
		if err := os.MkdirAll(OutputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory %q: %w", OutputDir, err)
		}
		tempDir, err := os.MkdirTemp("", "manga-tools-*")
		if err != nil {
			return fmt.Errorf("creating temp directory %q: %w", OutputDir, err)
		}
		defer os.RemoveAll(tempDir)

		var rootItem OutputUnit
		var items []OutputUnit

		if options.dir != "" {
			children, err := os.ReadDir(options.dir)
			if err != nil {
				return fmt.Errorf("listing directory %q contents: %w", options.dir, err)
			}
			children = internal.SortDirEntry(children)

			rootItem = OutputUnit{dir: options.dir}
			for _, child := range children {
				isAllowedFormat := slices.Contains(allowedFormats, filepath.Ext(strings.ToLower(child.Name())))
				if !child.IsDir() && !isAllowedFormat {
					continue
				}

				if !child.IsDir() && isAllowedFormat {
					rootItem.appendFile(filepath.Join(options.dir, child.Name()))
					continue
				}

				// child is directory
				subDir := filepath.Join(options.dir, child.Name())
				files, err := os.ReadDir(subDir)
				if err != nil {
					return fmt.Errorf("listing directory %q contents: %w", filepath.Join(options.dir, child.Name()), err)
				}
				files = internal.SortDirEntry(files)

				subDirOutput := OutputUnit{dir: subDir}
				for _, file := range files {
					isAllowedFormat := slices.Contains(allowedFormats, filepath.Ext(strings.ToLower(file.Name())))
					if (!file.IsDir() && !isAllowedFormat) || file.IsDir() {
						continue
					}
					subDirOutput.appendFile(filepath.Join(subDir, file.Name()))
				}
				if len(subDirOutput.files) == 0 {
					continue
				}
				items = append(items, subDirOutput)
			}
		}

		var images []model.FilePath
		for _, item := range items {
			result, err := item.getImages(tempDir, inputFormat)
			if err != nil {
				return fmt.Errorf("getting images for %q: %w", item.dir, err)
			}
			if options.bundle {
				images = append(images, result...)
				continue
			}

			outputFilePath := filepath.Join(OutputDir, fmt.Sprintf("%s.pdf", filepath.Base(item.dir)))
			if err := pdf.Save(outputFilePath, result); err != nil {
				return fmt.Errorf("saving pdf %q: %w", outputFilePath, err)
			}
		}

		if options.bundle || inputFormat == inputTypeImage {
			result, err := rootItem.getImages(tempDir, inputFormat)
			if err != nil {
				return fmt.Errorf("getting images for %q: %w", rootItem.dir, err)
			}
			images = append(images, result...)
			outputFilePath := filepath.Join(OutputDir, fmt.Sprintf("%s.pdf", filepath.Base(rootItem.dir)))
			if err := pdf.Save(outputFilePath, images); err != nil {
				return fmt.Errorf("saving pdf %q: %w", outputFilePath, err)
			}
			return nil
		}

		for _, archive := range rootItem.files {
			result, err := format.ExtractArchive(tempDir, rootItem.dir, filepath.Base(archive))
			if err != nil {
				return fmt.Errorf("extracting archive %q: %w", archive, err)
			}
			outputFilePath := filepath.Join(OutputDir, strings.Replace(filepath.Base(archive), filepath.Ext(archive), ".pdf", 1))
			if err := pdf.Save(outputFilePath, result); err != nil {
				return fmt.Errorf("saving pdf %q: %w", outputFilePath, err)
			}
		}

		return nil
	}
}
