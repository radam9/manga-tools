package cmd

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/radam9/manga-tools/internal"
	"github.com/radam9/manga-tools/internal/format"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

const (
	mergeDirFlag   = "dirs"
	mergeFilesFlag = "files"
)

type mergeOptions struct {
	dirs    []string
	files   []string
	archive bool
	image   bool
	bundle  bool
}

func NewMergeCommand() *cobra.Command {
	options := &mergeOptions{}

	cmd := &cobra.Command{
		Use:   "merge",
		Short: "merges a list of pdfs into a single file",
		Long: `merges a list of pdfs into a single file, given a list of directories and/or files.
The items are merged in the following order:
	1. Directories in the provided order, the files in the directory are sorted by filename.
	2. Files in the provided order.`,
		Args: cobra.NoArgs,
		RunE: mergeCommandRunFunction(options),
	}

	flags := cmd.Flags()
	flags.StringSliceP(mergeDirFlag, "d", nil, "comma separated list of path to directories containing pdf files to merge")
	flags.StringSliceP(mergeFilesFlag, "f", nil, "comma separated list of path to pdf files to merge")
	cmd.MarkFlagsOneRequired(mergeDirFlag, mergeFilesFlag)

	return cmd
}

func mergeCommandRunFunction(options *mergeOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var pdfs []string

		for _, dir := range options.dirs {
			children, err := os.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("listing files in directory %q: %w", dir, err)
			}
			children = internal.SortDirEntry(children)

			for _, child := range children {
				pdfPath := filepath.Join(dir, child.Name())
				if filepath.Ext(strings.ToLower(pdfPath)) != ".pdf" {
					continue
				}
				pdfs = append(pdfs, pdfPath)
			}
		}
		pdfs = append(pdfs, options.files...)

		_, conf := format.DefaultPDFConfig()

		if err := api.MergeCreateFile(pdfs, OutputDir, false, conf); err != nil {
			return fmt.Errorf("merging pdf files: %w", err)
		}
		return nil
	}
}
