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
	dir         string
	archiveMode bool
	imageMode   bool
	bundle      bool
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

	const convertArchiveModeFlag = "archive"
	const convertImageModeFlag = "image"
	flags.BoolVar(&options.archiveMode, convertArchiveModeFlag, false, "source files are archive format (cbr/cbz)")
	flags.BoolVar(&options.imageMode, convertImageModeFlag, false, "source files are image format")
	cmd.MarkFlagsMutuallyExclusive(convertArchiveModeFlag, convertImageModeFlag)

	flags.BoolVarP(&options.bundle, "bundle", "b", false, "bundle all passed dir and files into a single pdf")
	return cmd
}

func convertCommandRunFunction(options *convertOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(OutputDir, 0755); err != nil {
			return fmt.Errorf("creating output directory %q: %w", OutputDir, err)
		}
		tempDir, err := os.MkdirTemp("", "manga-tools-*")
		if err != nil {
			return fmt.Errorf("creating temp directory %q: %w", OutputDir, err)
		}
		defer os.RemoveAll(tempDir)

		rootName, items, err := parseDirs(options)
		if err != nil {
			return err
		}

		return writeOutputPDF(tempDir, items, rootName, options.imageMode, options.bundle)
	}
}

func parseDirs(options *convertOptions) (string, []convertOutputUnit, error) {
	allowedFormats := imageFormats
	if options.archiveMode {
		allowedFormats = archiveFormats
	}

	var rootItem convertOutputUnit
	var items []convertOutputUnit

	rootItem = convertOutputUnit{dir: options.dir, name: filepath.Base(options.dir)}
	children, err := os.ReadDir(options.dir)
	if err != nil {
		return "", nil, fmt.Errorf("listing directory %q contents: %w", options.dir, err)
	}
	children = internal.SortDirEntry(children)

	for _, child := range children {
		isAllowedFormat := slices.Contains(allowedFormats, filepath.Ext(strings.ToLower(child.Name())))
		if !child.IsDir() && !isAllowedFormat {
			continue
		}

		childPath := filepath.Join(options.dir, child.Name())
		if !child.IsDir() && isAllowedFormat {
			if (options.imageMode) || (options.archiveMode && options.bundle) {
				rootItem.appendFile(childPath)
			} else if options.archiveMode && !options.bundle {
				item := convertOutputUnit{
					dir:   options.dir,
					name:  child.Name(),
					files: []model.FilePath{childPath},
				}
				items = append(items, item)
			}
			continue
		}

		// child is directory
		files, err := os.ReadDir(childPath)
		if err != nil {
			return "", nil, fmt.Errorf("listing directory %q contents: %w", childPath, err)
		}
		files = internal.SortDirEntry(files)

		subDirOutput := convertOutputUnit{dir: childPath, name: filepath.Base(child.Name())}
		for _, file := range files {
			isAllowedFormat := slices.Contains(allowedFormats, filepath.Ext(strings.ToLower(file.Name())))
			if (!file.IsDir() && !isAllowedFormat) || file.IsDir() {
				continue
			}
			subDirOutput.appendFile(filepath.Join(childPath, file.Name()))
		}
		items = append(items, subDirOutput)
	}
	items = append(items, rootItem)

	return rootItem.name, items, nil
}

func writeOutputPDF(tempDir string, items []convertOutputUnit, rootName string, imageMode, bundle bool) error {
	pdf := format.PDF{}

	var images []model.FilePath
	for _, item := range items {
		result, err := item.getImages(tempDir, imageMode)
		if err != nil {
			return fmt.Errorf("getting images for %q: %w", item.dir, err)
		}
		if bundle {
			images = append(images, result...)
			continue
		}

		outputFilePath := filepath.Join(OutputDir, fmt.Sprintf("%s.pdf", item.name))
		if err := pdf.Save(outputFilePath, result); err != nil {
			return fmt.Errorf("saving pdf %q: %w", outputFilePath, err)
		}
	}

	if bundle {
		outputFilePath := filepath.Join(OutputDir, fmt.Sprintf("%s.pdf", rootName))
		if err := pdf.Save(outputFilePath, images); err != nil {
			return fmt.Errorf("saving pdf %q: %w", outputFilePath, err)
		}
	}
	return nil
}

type convertOutputUnit struct {
	dir   string
	name  string
	files []model.FilePath
}

func (c convertOutputUnit) getImages(tempDir string, imageMode bool) ([]string, error) {
	var images []string
	if imageMode {
		for _, image := range c.files {
			images = append(images, image)
		}
		return images, nil
	}

	// archive mode
	for _, archive := range c.files {
		imagePaths, err := format.ExtractArchive(tempDir, c.dir, filepath.Base(archive))
		if err != nil {
			return nil, fmt.Errorf("extracting archive %q: %w", archive, err)
		}
		images = append(images, imagePaths...)
	}
	return images, nil
}

func (c *convertOutputUnit) appendFile(newFile model.FilePath) {
	currentFiles := c.files
	currentFiles = append(currentFiles, newFile)
	c.files = currentFiles
}
