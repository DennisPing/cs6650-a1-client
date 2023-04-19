.PHONY: all httpclient clean

all: httpclient

httpclient:
	go build -o httpclient .

clean:
	rm -f httpclient
