.PHONY: all
all: clean test install

.PHONY: clean
clean:
	mkdir -p public/
	rm -rf public/*

.PHONY: test
test:
		go test

.PHONY: install
install: 
		go install .