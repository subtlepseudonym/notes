BINARY=notes

BUILD=$$(bash build-tag.sh)
REVISION=`git rev-list -n1 HEAD`
LDFLAGS=--ldflags "-X main.Version=${BUILD} -X main.Revision=${REVISION}"

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
