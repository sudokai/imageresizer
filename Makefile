BINARY_NAME=imageresizer

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get -u

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
