BINARY_NAME=novaexample

.PHONY: build clean fmt help
default: build

build:
	@go build -o $(BINARY_NAME)

clean:
	@rm -f $(BINARY_NAME)

fmt:
	@goimports -w .
	@go fmt ./...

help:
	@echo "Available Make targets:"
	@echo "  build : Build the Go application (default)"
	@echo "  clean : Remove the built binary ($(BINARY_NAME))"
	@echo "  fmt   : Format Go source code (using goimports)"
	@echo "  help  : Show this help message"
