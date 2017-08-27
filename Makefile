.phony: all run clean watch

BINARY=hanabi-server

all: hanabi-server

$(BINARY):
	go build

run: $(BINARY)
	./$(BINARY)

watch: $(BINARY)
	ls *.go | entr -r sh -c "make clean && make run"

clean:
	rm -f ./$(BINARY)
