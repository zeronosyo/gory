help:
	@echo "help:"
	@echo "  make build-dev   - build gory binary file of current platform"
	@echo "  make build-prod  - build gory binary file of linux platform"
	@echo "  make clean       - clean file generate by `make build`"
	@echo "  make lint        - lint all go source code by a lot linters"
	@echo "  make deps        - install depends of gory"
	@echo "  make test        - run tests"

lint:
	gometalinter \
		--enable-all \
		--fast \
		--errors \
		--enable=safesql \
		--disable=gosec \
		--aggregate \
		--vendor \
		./...

deps:
	dep ensure

build-prod: deps
	GOOS=linux GOARCH=amd64 go build -o gory-prod

build-dev: deps
	go build -o gory-dev

clean:
	rm -rf ./.vendor-new
	rm -rf ./gory-dev
	rm -rf ./gory-prod
	go clean

test: deps lint build-dev build-prod

.PHONY: build-prod build-dev lint clean deps
