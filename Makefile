BIN_UI = fuzzerUI
BIN_CLI = fuzzerCLI

INSTALL_PATH=/usr/local/bin

.PHONY: all tidy build install clean

all: tidy build

# Ensures all dependencies are downloaded and go.mod is synchronized
tidy:
	go mod tidy

# Standard build command
build:
	go build -o $(BIN_UI) ./cmd/fuzzerUI
	go build -o $(BIN_CLI) ./cmd/fuzzerCLI
	
# Install requires a build (which now runs tidy first)
install: build
	sudo cp $(BIN_UI) $(BIN_CLI) $(INSTALL_PATH)/
	sudo chmod +x $(INSTALL_PATH)/$(BIN_UI) $(INSTALL_PATH)/$(BIN_CLI)
	@echo "Installed $(BIN_UI) and $(BIN_CLI) to $(INSTALL_PATH)"

clean:
	rm -f $(BIN_UI) $(BIN_CLI)
