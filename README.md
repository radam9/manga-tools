# manga-tools

A set of tools for reading mangas.

## Features:

- Downloader
  - Download mangas from MangaDex as image, cbr, cbz, or pdf
- Converter
  - cbr to pdf
  - cbz to pdf
  - images (jpg/jpeg/png) to pdf
- Merger
  - merge pdfs into a single pdf file

## Install:
There are differnet ways to install the tool:
1. Download the already built executables in the release section.
2. Install using the go toolchain `go install github.com/radam9/manga-tools`
3. Download the source code and build:
```bash
# using http "https://github.com/radam9/manga-tools.git"
git clone git@github.com:radam9/manga-tools
cd manga-tools
make build
```

## Usage:
```terminaloutput
manga-tools allows you to download mangas from mangadex as CBZ or PDF.
            it also allows the conversion of mangas from cbz or images to pdf, and combining pdfs into one file.

Usage:
  manga-tools [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  convert     convert a set of images, cbr or cbz files to pdf
  download    downloads a manga from mangadex given a url/id
  help        Help about any command
  merge       merges a list of pdfs into a single file
  version     Print the version number of manga-tools

Flags:
  -h, --help            help for manga-tools
  -o, --output string   path to output directory (default is current directory)

Use "manga-tools [command] --help" for more information about a command.
```
