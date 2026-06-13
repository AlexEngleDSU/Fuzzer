# Name of the single consolidated binary
BIN_NAME = fuzzer

INSTALL_PATH = /usr/local/bin

.PHONY: all tidy build install clean

all: tidy build

# Ensures all dependencies are downloaded and go.mod is synchronized
tidy:
	go mod tidy

# Build the single binary from the new router location
build:
	go build -o $(BIN_NAME) ./cmd/fuzzer

# Install the single binary
install: build
	sudo cp $(BIN_NAME) $(INSTALL_PATH)/
	sudo chmod +x $(INSTALL_PATH)/$(BIN_NAME)
	@echo "Installed $(BIN_NAME) to $(INSTALL_PATH)"

# Clean the single binary
clean:
	rm -f $(BIN_NAME)
