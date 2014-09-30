all: maud

maud:
	go build -o maud ./src

run: all
	./maud

clean:
	rm -f maud