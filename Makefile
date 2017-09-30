all: maud

maud: grunt dep core

.PHONY: core
core:
	go install github.com/hamcha/maud/maud

grunt:
	grunt build
	touch grunt

run: all
	maud

dep:
	go get github.com/gorilla/mux
	go get github.com/microcosm-cc/bluemonday
	go get github.com/oschwald/maxminddb-golang
	go get github.com/bamiaux/rez
	go get gopkg.in/mgo.v2
	touch dep

clean:
	rm -f dep grunt

test:
	go test ./...
