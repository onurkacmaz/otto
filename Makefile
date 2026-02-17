BINARY_NAME=otto
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: build install uninstall

build:
	go build -o $(BINARY_NAME) .

install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) installed: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "  If not in PATH, add: export PATH=\"$$HOME/.local/bin:$$PATH\""

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) removed"
