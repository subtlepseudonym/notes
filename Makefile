GOBINARY=notes

all: test build

build:
	go build -o $(GOBINARY) -v

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
