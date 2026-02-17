BINARY_NAME=db-console
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: build install uninstall

build:
	go build -o $(BINARY_NAME) .

install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) kuruldu: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "  PATH'te yoksa şunu ekle: export PATH=\"$$HOME/.local/bin:$$PATH\""

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) kaldırıldı"
