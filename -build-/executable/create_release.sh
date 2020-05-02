#!/bin/sh

echo "Pruning the old releases..."
rm -R releases

echo "Creating folders..."
mkdir releases
mkdir releases/linux
mkdir releases/macosx
mkdir releases/windows

major=$(date +%y)
buildNo=$(($(date +%s)/345600))

export RELEASE_VERSION="$major.1.$buildNo"

echo ""
echo "Building Locking Center (v$RELEASE_VERSION)"
cd ../../mutex
echo "  > compiling linux x64 release"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/linux/locking-center
echo "  > compiling macosx x64 release"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/macosx/locking-center
echo "  > compiling windows x64 release"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/windows/locking-center.exe
