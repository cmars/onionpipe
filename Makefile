
all: onionpipe

onionpipe_libtor:
	go build -o $@ -v -x -tags "staticOpenssl,staticZlib,staticLibevent,libtor" .
	strip $@

onionpipe_embed:
	go build -o $@ -v -x -tags "embed" .

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
