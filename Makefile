all: maud

maud: dep
	go build -o maud ./src

run: all
	./maud

dep:
	go get github.com/gorilla/mux
	go get github.com/hoisie/mustache
	go get gopkg.in/mgo.v2
	touch dep

clean:
	rm -f maud