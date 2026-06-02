BINARY_NAME=fuzzer
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)

.PHONY: all tidy build install clean

all: tidy build

# Ensures all dependencies are downloaded and go.mod is synchronized
tidy:
	go mod tidy

# Standard build command
build:
	go build -o fuzzer .

# Install requires a build (which now runs tidy first)
install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)
	@echo "Installed to $(INSTALL_PATH)"

clean:
	rm -f $(BINARY_NAME)
