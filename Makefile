build:
	go build -o bin/scraper ./cmd/scraper

run:
	go run ./cmd/scraper/main.go

clean:
	rm -rf bin/
