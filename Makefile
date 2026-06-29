BINARY=run-cmd
GOBUILD=go build -ldflags="-s -w" -o $(BINARY)

.PHONY: all build clean install run

all: build

build:
	$(GOBUILD) .

clean:
	rm -f $(BINARY)

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

run: build
	./$(BINARY) "$(CMD)"
