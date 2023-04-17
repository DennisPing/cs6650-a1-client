.PHONY: all client clean

all: client

client:
	go build -o ./client

clean:
	rm -f ./client
