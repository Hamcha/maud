all: maud

maud: grunt dep
	go build -o maud ./src

grunt:
	grunt build

run: all
	./maud

dep:
	go get github.com/gorilla/mux
	go get github.com/hoisie/mustache
	go get gopkg.in/mgo.v2
	touch dep

clean:
	rm -f maud dep