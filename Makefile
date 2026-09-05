GO         ?= go
GOLANGCI   ?= golangci-lint
COVERPROFILE ?= coverage.out

.PHONY: all
all: check

## check: run everything the CI runs
.PHONY: check
check: tidy-check fmt-check vet lint test

.PHONY: build
build:
	$(GO) build ./...

.PHONY: test
test:
	$(GO) test -race -shuffle=on ./...

.PHONY: cover
cover:
	$(GO) test -coverprofile=$(COVERPROFILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERPROFILE)

.PHONY: cover-html
cover-html: cover
	$(GO) tool cover -html=$(COVERPROFILE)

.PHONY: bench
bench:
	$(GO) test -run '^$$' -bench . -benchmem ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: fmt-check
fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt'd:"; echo "$$out"; exit 1; fi

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	$(GOLANGCI) run ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

# Checks tidiness against the files on disk rather than against git, so it
# gives the same answer with a dirty working tree.
.PHONY: tidy-check
tidy-check:
	@cp go.mod go.mod.bak
	@[ -f go.sum ] && cp go.sum go.sum.bak || true
	@$(GO) mod tidy
	@status=0; \
	cmp -s go.mod go.mod.bak || status=1; \
	if [ -f go.sum.bak ] || [ -f go.sum ]; then cmp -s go.sum go.sum.bak || status=1; fi; \
	mv go.mod.bak go.mod; \
	[ -f go.sum.bak ] && mv go.sum.bak go.sum || true; \
	if [ $$status -ne 0 ]; then echo "go.mod/go.sum are not tidy; run make tidy"; fi; \
	exit $$status

.PHONY: doc
doc:
	$(GO) run golang.org/x/pkgsite/cmd/pkgsite@latest -open .

.PHONY: clean
clean:
	rm -f $(COVERPROFILE)

.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
	@grep -E '^\.PHONY: ' $(MAKEFILE_LIST) | sed 's/\.PHONY: /  make /'
