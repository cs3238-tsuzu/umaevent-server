.PHONY: umaevent-server run generate
umaevent-server: generate
	go build -tags="$(TAG)" -o umaevent-server .

generate:
	go generate -tags="$(TAG)" -v ./...

run: generate
	go run -tags="$(TAG)" .
