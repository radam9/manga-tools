package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var Version = "source"
var OutputDir string

var rootCmd = &cobra.Command{
	Use:   "manga-tools [command]",
	Short: "manga-tools are a set of tools to download, convert and manipulate mangas.",
	Long: `manga-tools allows you to download mangas from mangadex as CBZ or PDF.
it also allows the conversion of mangas from cbz or images to pdf, and combining pdfs into one file.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of manga-tools",
	Long:  `Print the version number of manga-tools`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("manga-tools %s\n", Version)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&OutputDir, "output", "o", "", "path to output directory (default is current directory)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(NewDownloadCommand())
	rootCmd.AddCommand(NewConvertCommand())
	rootCmd.AddCommand(NewMergeCommand())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("failed to execute the cmd", "error", err)
		os.Exit(1)
	}
}
