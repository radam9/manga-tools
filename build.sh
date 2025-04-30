#!/usr/bin/env bash

version=$1
if [[ -z "$version" ]]; then
  echo "usage: $0 <version>"
  exit 1
fi

package="github.com/radam9/manga-tools"
package_split=(${package//\// })
package_name=${package_split[-1]}

platforms=(
  "windows/386"
  "windows/amd64"
  "linux/386"
  "linux/amd64"
  "linux/arm"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$version'-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi

	env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done