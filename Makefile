all: maud

maud: grunt dep
	go build -o maud ./src

grunt:
	grunt build
	touch grunt

run: all
	./maud

dep:
	go get github.com/gorilla/mux
	go get github.com/microcosm-cc/bluemonday
	go get github.com/oschwald/maxminddb-golang
	go get github.com/bamiaux/rez
	go get gopkg.in/mgo.v2
	touch dep

clean:
	rm -f maud dep grunt

go-clean:
	rm -f maud
