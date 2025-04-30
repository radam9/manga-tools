version=$(shell git describe --tags --abbrev=0)
build:
	go build -ldflags "-X github.com/radam9/manga-tools/cmd.Version=$(version)" .

build-platforms:
	./build.sh $(version)
