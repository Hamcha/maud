all: maud

maud: grunt core

.PHONY: core
core:
	go install ./maud

grunt:
	grunt build
	touch grunt

run: all
	maud

clean:
	rm -f grunt

test:
	go test ./...
