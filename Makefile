PREFIX ?= $(HOME)/.local/bin
BINARY  = switcher

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(PREFIX)
	install -m 755 $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	rm -f $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
