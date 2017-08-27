.phony: all run clean watch

BINARY=hanabi-server

all: hanabi-server

$(BINARY):
	go build

run: $(BINARY)
	./$(BINARY)

watch:
	ls *.go | entr -r sh -c "make run"

clean:
	rm -f ./$(BINARY)
