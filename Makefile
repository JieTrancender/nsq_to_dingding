BUILDDIR = build
BUILDFLAGS =

APPS = nsq_to_dingding
all: $(APPS)

$(BUILDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BUILDFLAGS} -o $@ ./

$(APPS): %: $(BUILDDIR)/%

clean:
	rm -rf $(BUILDDIR)

test:
	go test -v -race -cover -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: clean all test lint
.PHONY: $(APPS)

lint:
	golangci-lint cache clean
	golangci-lint run --tests=false ./...
