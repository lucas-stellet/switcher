PREFIX ?= /usr/local/bin
BINARY  = switcher

build:
	go build -o $(BINARY) .

install: build
	sudo install -m 755 $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	sudo rm -f $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
