BINARY=notes
VERSION=`git describe --abbrev=0`
LDFLAGS=--ldflags "-X main.Version=${VERSION}"

all: test build

build: format
	go build ${LDFLAGS} -o ${BINARY} -v ./cmd/notes

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
