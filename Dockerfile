FROM golang:buster
WORKDIR /workspace/

RUN apt-get update
RUN go get -u github.com/subtlepseudonym/utilities/cmd/vtag
RUN apt-get install --assume-yes \
	git \
	vim

COPY . .
RUN make dev-build

CMD ["/bin/bash"]
