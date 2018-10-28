BINARY_NAME=imageresizer

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

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
dep:
	$(GOGET) github.com/aws/aws-sdk-go
	$(GOGET) github.com/cloudflare/tableflip
	$(GOGET) github.com/djherbis/atime
	$(GOGET) github.com/gorilla/mux
	$(GOGET) github.com/pkg/errors
	$(GOGET) github.com/rcrowley/go-metrics
	$(GOGET) github.com/spf13/viper