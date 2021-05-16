.PHONY: umaevent-server
umaevent-server:
	go build -o umaevent-server .

run: umaevent-server
	./umaevent-server
