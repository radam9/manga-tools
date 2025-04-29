package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/radam9/manga-tools/internal/format"
	"github.com/radam9/manga-tools/internal/mangadex"
	"github.com/radam9/manga-tools/internal/model"
	"github.com/radam9/manga-tools/internal/ranges"
	"github.com/spf13/cobra"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	maxChapterConcurrency = 5
	maxPageConcurrency    = 10
)

type DownloadOptions struct {
	language     string
	bundle       bool
	bundleVolume bool
	chapterRange string
	image        bool
	cbr          bool
	cbz          bool
	pdf          bool
}

func NewDownloadCommand() *cobra.Command {
	options := &DownloadOptions{}

	cmd := &cobra.Command{
		Use:   "download URL/ID",
		Short: "downloads a manga from mangadex given a url/id",
		Long: `downloads a manga from mangadex given the url or id of the manga.
By default the manga is downloaded as image files, specify the appropriate flag to download in a different format.`,
		Example: `Download entire manga using url
	$ manga-tools download https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk

Download entire manga using id
	$ manga-tools download 319df2e2-e6a6-4e3a-a31c-68539c140a84

Download manga as pdf
	$ manga-tools download https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk --pdf

Download manga as cbz and bundle into one file
	$ manga-tools download https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk --bundle --cbz

Download range of chapters
	$ manga-tools download https://mangadex.org/title/319df2e2-e6a6-4e3a-a31c-68539c140a84/slam-dunk -c 1-20

Ranges can take the following forms:
	- 1-20
	- 1,2,5-10
	- 1-20.5 (include decimal values)`,
		Args: cobra.MinimumNArgs(1),
		RunE: downloadCommandRunFunction(options),
	}

	const downloadBundleFlag = "bundle"
	const downloadBundleVolumeFlag = "bundle-volume"

	flags := cmd.Flags()
	flags.StringVarP(&options.language, "language", "l", "", "the manga language to download")
	flags.BoolVarP(&options.bundle, downloadBundleFlag, "b", false, "bundle all downloads into a single file (single folder for images)")
	flags.BoolVarP(&options.bundleVolume, downloadBundleVolumeFlag, "B", false, "bundle all downloads into a single file (single folder for images) per volume")
	cmd.MarkFlagsMutuallyExclusive(downloadBundleFlag, downloadBundleVolumeFlag)

	flags.StringVarP(&options.chapterRange, "chapters", "c", "", "chapter range to download")

	const downloadImageFormatFlag = "image"
	const downloadCBRFormatFlag = "cbr"
	const downloadCBZFormatFlag = "cbz"
	const downloadPDFFormatFlag = "pdf"
	flags.BoolVar(&options.image, downloadImageFormatFlag, false, "download manga in image format")
	flags.BoolVar(&options.image, downloadCBRFormatFlag, false, "download manga in CBR format")
	flags.BoolVar(&options.image, downloadCBZFormatFlag, false, "download manga in CBZ format")
	flags.BoolVar(&options.image, downloadPDFFormatFlag, false, "download manga in PDF format")
	cmd.MarkFlagsMutuallyExclusive(downloadImageFormatFlag, downloadCBRFormatFlag, downloadCBZFormatFlag, downloadPDFFormatFlag)
	return cmd
}

func downloadCommandRunFunction(options *DownloadOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		mangaID, err := parseURLOrID(args[0])
		if err != nil {
			return err
		}

		saver := format.SelectFormat(options.cbr, options.cbz, options.pdf)

		client := mangadex.NewClient(mangaID, options.language)
		mangaTitle, err := client.FetchTitle()
		if err != nil {
			slog.Error("fetching manga title", "error", err)
			return err
		}

		chapters, errs := client.FetchChapterList()
		if len(errs) > 0 {
			slog.Error("fetching manga chapters", "title", mangaTitle, "errors", errs)
			os.Exit(1)
		}

		model.SortChaptersByNumber(chapters)

		if options.chapterRange == "" {
			fmt.Printf("Do you want to download all %g chapters? [y/n]: ", chapters[len(chapters)-1].Number)
			var downloadAll string
			_, err := fmt.Scan(&downloadAll)
			if err != nil || !strings.EqualFold(downloadAll, "y") {
				os.Exit(0)
			}
			options.chapterRange = fmt.Sprintf("%f-%f", chapters[0].Number, chapters[len(chapters)-1].Number)
		}
		chapterRanges, err := ranges.Parse(options.chapterRange)
		if err != nil {
			slog.Error("parsing chapter range", "error", err)
			os.Exit(1)
		}

		chapters = model.FilterChapters(chapters, chapterRanges)
		if len(chapters) == 0 {
			slog.Info("no manga chapters found")
			os.Exit(0)
		}

		tempDir, err := os.MkdirTemp("", "")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		wg := sync.WaitGroup{}
		guard := make(chan struct{}, maxChapterConcurrency)

		for i := range len(chapters) {
			guard <- struct{}{}
			wg.Add(1)
			go func(chapter *model.Chapter) {
				defer wg.Done()

				slog.Info("fetching chapter", "chapterID", chapter.ID, "chapterTitle", chapter.Title)
				err = client.FetchChapterInfo(chapter)
				if err != nil {
					slog.Error("fetching chapter", "chapterID", chapter.ID, "chapterTitle", chapter.Title, "error", err)
					<-guard
					return
				}

				slog.Info("downloading chapter pages", "chapterID", chapter.ID, "chapterTitle", chapter.Title)
				pages, err := client.FetchChapterPages(chapter.Number, chapter.ID, chapter.Pages, maxPageConcurrency)
				if err != nil {
					slog.Error("downloading chapter pages", "chapterID", chapter.ID, "chapterTitle", chapter.Title, "error", err)
					<-guard
					return
				}
				if len(pages) == 0 {
					<-guard
					return
				}
				if chapter.Pages, err = mangadex.WritePagesToTempFiles(tempDir, pages); err != nil {
					slog.Error("writing page to temp files", "error", err)
				}

				if !options.bundle && !options.bundleVolume {
					filename := saver.OutputPath(OutputDir, mangaTitle, chapter.Volume, chapter.Title, chapter.Number)
					slog.Info("writing output file", "filepath", filename)
					if err := saver.Save(filename, model.GetSliceOfPagePathsFromPages(pages)); err != nil {
						slog.Error("saving chapter", "filename", filename, "error", err)
					}
				}

				<-guard
				return
			}(&chapters[i])
		}
		wg.Wait()
		close(guard)

		if options.bundle {
			filename := saver.OutputPath(OutputDir, mangaTitle, 0, "", 0)
			var pagesFilePaths []model.FilePath
			for _, chapter := range chapters {
				if len(chapter.Pages) == 0 {
					continue
				}
				pagesFilePaths = append(pagesFilePaths, model.GetSliceOfPagePathsFromPages(chapter.Pages)...)
			}

			slog.Info("writing output file", "filepath", filename)
			if err = saver.Save(filename, pagesFilePaths); err != nil {
				return fmt.Errorf("writing pages to bundle pdf file: %w", err)
			}
		}

		if options.bundleVolume {
			volumes := map[string][]model.FilePath{}
			for _, chapter := range chapters {
				if len(chapter.Pages) == 0 {
					continue
				}
				filename := saver.OutputPath(OutputDir, mangaTitle, chapter.Volume, "", 0)
				volumes[filename] = append(volumes[filename], model.GetSliceOfPagePathsFromPages(chapter.Pages)...)
			}

			for filename, pages := range volumes {
				slog.Info("writing output file", "filepath", filename)
				if err := saver.Save(filename, pages); err != nil {
					slog.Error("writing volume to pdf file", "filename", filename, "error", err)
				}
			}
		}

		return nil
	}
}

func parseURLOrID(s string) (uuid.UUID, error) {
	u, err := url.Parse(s)
	if err != nil {
		return uuid.Parse(s)
	}
	for _, part := range strings.Split(u.Path, "/") {
		if id, err := uuid.Parse(part); err == nil {
			return id, nil
		}
	}
	return uuid.Nil, fmt.Errorf("could not parse ID %s", s)
}
