
all: oniongrok

oniongrok_libtor:
	go build -o $@ -v -x -tags "staticOpenssl,staticZlib,staticLibevent,libtor" .
	strip $@

oniongrok_embed:
	go build -o $@ -v -x -tags "embed" .

oniongrok:
	go build -o $@ .

.PHONY: docker
docker:
	docker build -t oniongrok .

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
	$(RM) oniongrok
