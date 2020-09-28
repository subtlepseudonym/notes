BINARY=notes

BUILD=$$(vtag)

REVISION=`git rev-list -n1 HEAD`
BUILDTAGS=
LDFLAGS=--ldflags "-X main.Version=${BUILD} -X main.Revision=${REVISION} -X \"main.BuildTags=${BUILDTAGS}\""

DEV_BUILDTAGS= debug
space=$(eval) #
comma=,

default: all

all: test build

build: format
	go build ${LDFLAGS} -o ${BINARY} -v ./cmd/notes

dev: BUILDTAGS=$(subst $(space),$(comma),$(DEV_BUILDTAGS))
dev: format
	go build -tags "${BUILDTAGS}" ${LDFLAGS} -o ${BINARY} -v ./cmd/notes

test:
	gotest --race -v ./...

dev-test: BUILDTAGS=$(subst $(space),$(comma),$(DEV_BUILDTAGS))
dev-test:
	gotest --race --count=1 -v -tags "${BUILDTAGS}" ./...

format fmt:
	gofmt -l -w .

clean:
	go mod tidy
	go clean
	rm $(BINARY)

get-tag:
	echo ${BUILD}

.PHONY: all build dev-build test format fmt clean get-tag
