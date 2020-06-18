.PHONY: build build-environment-build lambda-environment-build

dist.zip: build
	zip -j dist.zip build/main

build:
	go generate ../...
	docker-compose run --rm build-environment make build-environment-build
	docker-compose run --rm lambda-environment make lambda-environment-build

build-environment-build:
	rm -rf ./build
	go build -o ./build/main .
	./build/main --help

lambda-environment-build:
	./build/main --help || (ldd ./build/main && exit 1)
