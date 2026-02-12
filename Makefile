PREFIX ?= /usr/local/bin
BINARY  = switch

build:
	go build -o $(BINARY) .

install: build
	cp $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	rm -f $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY)

.PHONY: build install uninstall clean
