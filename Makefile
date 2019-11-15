BINARY=notes

BUILD=$$(vtag)

REVISION=`git rev-list -n1 HEAD`
BUILDTAGS=
LDFLAGS=--ldflags "-X main.Version=${BUILD} -X main.Revision=${REVISION} -X \"main.BuildTags=${BUILDTAGS}\""

all: test build

build: format
	go build ${LDFLAGS} -o ${BINARY} -v ./cmd/notes

dev-build: BUILDTAGS=debug
dev-build: format
	go build -tags "${BUILDTAGS}" ${LDFLAGS} -o ${BINARY} -v ./cmd/notes

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

get-tag:
	echo ${BUILD}
