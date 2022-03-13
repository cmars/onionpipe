
all: onionpipe

onionpipe:
	go build -o $@ .

.PHONY: docker
docker:
	docker build -t onionpipe .

.PHONY: test
test:
	go test ./... -count=1

.PHONY: coverage
coverage:
	go test ./... -count=1 -coverprofile=covfile
	go tool cover -html=covfile
	rm -f covfile

.PHONY: clean
clean:
	$(RM) onionpipe
