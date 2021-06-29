#!/bin/sh

echo "Pruning the old releases..."
rm -R releases

echo "Creating folders..."
mkdir releases
mkdir -p releases/linux/arm64
mkdir -p releases/linux/amd64
mkdir -p releases/macosx/arm64
mkdir -p releases/macosx/amd64
mkdir -p releases/windows/arm64
mkdir -p releases/windows/amd64

major=$(date +%y)
buildNo=`printf %04d $(expr $(expr $(date +%s) - $(gdate -d "Jul 2 2020" +%s)) / 345600)`
export RELEASE_VERSION="$major.2.$buildNo"
export BUILD=`printf %04d $(expr $(expr $(date +%s) - $(gdate -d "Jun 13 2020" +%s)) / 96)`

go clean -cache

echo ""
echo "Building Locking Center Server (v$RELEASE_VERSION)"
cd ../../mutex
echo "  > compiling linux arm64 release"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/linux/arm64/lcd
echo "  > compiling linux amd64 release"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/linux/amd64/lcd
echo "  > compiling macosx arm64 release"
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/macosx/arm64/lcd
echo "  > compiling macosx amd64 release"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/macosx/amd64/lcd
echo "  > compiling windows amd64 release"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/windows/amd64/lcd.exe

echo ""
echo "Building Locking Center CLI (v$RELEASE_VERSION)"
cd ../cli
echo "  > compiling linux arm64 release"
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/linux/arm64/lc-cli
echo "  > compiling linux amd64 release"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/linux/amd64/lc-cli
echo "  > compiling macosx arm64 release"
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/macosx/arm64/lc-cli
echo "  > compiling macosx amd64 release"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/macosx/amd64/lc-cli
echo "  > compiling windows amd64 release"
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$RELEASE_VERSION -X main.build=$BUILD" -o ../-build-/executable/releases/windows/amd64/lc-cli.exe

