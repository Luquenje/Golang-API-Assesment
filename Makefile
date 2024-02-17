build:	
	go build -o bin/program

run : build
	./bin/program

test:
	go test -v ./...

get:
	go get -d -v ./...