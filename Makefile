
all: oniongrok

oniongrok:
	go build -v -x -tags "staticOpenssl,staticZlib,staticLibevent" .

.PHONY: docker
docker:
	docker build -t oniongrok .

.PHONY: clean
clean:
	$(RM) oniongrok
