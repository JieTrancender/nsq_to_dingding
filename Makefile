BUILDDIR = build
BUILDFLAGS =

APPS = atomic-test nsq-producer nsq-consumer cobra-test zap-test
all: $(APPS)

$(BUILDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BUILDFLAGS} -o $@ ./apps/$*

$(APPS): %: $(BUILDDIR)/%

clean:
	rm -rf $(BUILDDIR)

test:
	go test -v -race -cover -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: clean all test lint
.PHONY: $(APPS)

lint:
	golangci-lint run --tests=false ./...