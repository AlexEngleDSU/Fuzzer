BINARY_NAME=fuzzer
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)

build:
	go build -o $(BINARY_NAME) ./cmd/fuzzer/main.go

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)
	@echo "Installed to $(INSTALL_PATH)"

clean:
	rm -f $(BINARY_NAME)
