all: binary container clean

binary:
	go get -d -v ./...
	CGO_ENABLED=0 go build -o main .

container:
	sudo docker build -t detection:latest .

clean:
	rm ./main
