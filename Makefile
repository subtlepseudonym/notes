BINARY=notes
VERSION=`git describe --abbrev=0`
BUILD=$$(cat build)
LDFLAGS=--ldflags "-X main.Version=${VERSION}+${BUILD}"

NEW_BUILD=$$(($(BUILD) + 1))

all: test build

build: format
	go build ${LDFLAGS} -o ${BINARY} -v ./cmd/notes
	echo $(NEW_BUILD) > ./build

test:
	gotest --race -v ./...

test-all:
	gotest --race --count=1 -v ./...

format fmt:
	go fmt -x ./...

clean:
	go mod tidy
	go clean
	rm -f $(GOBINARY)
