
all: oniongrok

oniongrok:
	go build -v -x -tags "staticOpenssl,staticZlib,staticLibevent" .
	strip oniongrok

.PHONY: docker
docker:
	docker build -t oniongrok .

.PHONY: test
test:
	go test ./... -count=1

.PHONY: clean
clean:
	$(RM) oniongrok
