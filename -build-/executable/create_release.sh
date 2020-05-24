#!/bin/sh

echo "Pruning the old releases..."
rm -R releases

echo "Creating folders..."
mkdir releases
mkdir releases/linux
mkdir releases/macosx
mkdir releases/windows

buildNo=$(($(date +%s)/345600))

export RELEASE_VERSION="0.1.$buildNo"

echo ""
echo "Building Locking Center Server (v$RELEASE_VERSION)"
cd ../../mutex
echo "  > compiling linux x64 release"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/linux/lcd
echo "  > compiling macosx x64 release"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/macosx/lcd
echo "  > compiling windows x64 release"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/windows/lcd.exe

echo ""
echo "Building Locking Center CLI (v$RELEASE_VERSION)"
cd ../cli
echo "  > compiling linux x64 release"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/linux/lc-cli
echo "  > compiling macosx x64 release"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/macosx/lc-cli
echo "  > compiling windows x64 release"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION" -o ../-build-/executable/releases/windows/lc-cli.exe
