BIN      = $(CURDIR)/bin
MODULE   = $(shell env GO111MODULE=on $(GO) list -m)
APP_NAME = $(notdir $(MODULE))
REGISTRY = harbor.ecn.zenlayer.net/monitoring
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
TESTPKGS = $(shell env GO111MODULE=on $(GO) list ./...)
OS     := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH   := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOPATH  = $(shell go env GOPATH)
GOBIN   = $(GOPATH)/bin
GO      = go
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

export GO111MODULE=on
export CGO_ENABLED=0

.PHONY: all
all: tests build ## Run tests and build

.PHONY: build
build: fmt lint | $(BIN) ; $(info $(M) building executable…) @ ## Build program binary
	$Q $(GO) build \
		-tags release \
		-ldflags "-X $(MODULE)/pkg/version.VERSION=$(VERSION) -X $(MODULE)/pkg/version.BUILDDATE=$(DATE)" \
		-o $(BIN)/$(notdir $(MODULE)) cmd/server/main.go

.PHONY: build-linux
build-linux: fmt lint | $(BIN) ; $(info $(M) building linux executable…) @ ## Build program binary with Linux
	$Q GOOS=linux GOARCH=amd64 $(GO) build \
		-tags release \
		-ldflags "-X $(MODULE)/pkg/version.VERSION=$(VERSION) -X $(MODULE)/pkg/version.BUILDDATE=$(DATE)" \
		-o $(BIN)/$(notdir $(MODULE)) cmd/server/main.go

# Tools

$(BIN):
	@mkdir -p $@
$(GOBIN):
	@mkdir -p $@
$(GOBIN)/%: | $(GOBIN) ; $(info $(M) building $(PACKAGE)…)
	$Q tmp=$$(mktemp -d); \
	   env GO111MODULE=off GOPATH=$$tmp GOBIN=$(GOBIN) $(GO) get $(PACKAGE) \
		|| ret=$$?; \
	   rm -rf $$tmp ; exit $$ret

GOLINT = $(GOBIN)/golint
$(GOBIN)/golint: PACKAGE=golang.org/x/lint/golint

GOCOV = $(GOBIN)/gocov
$(GOBIN)/gocov: PACKAGE=github.com/axw/gocov/...

GOCOVXML = $(GOBIN)/gocov-xml
$(GOBIN)/gocov-xml: PACKAGE=github.com/AlekSi/gocov-xml

GO2XUNIT = $(GOBIN)/go2xunit
$(GOBIN)/go2xunit: PACKAGE=github.com/tebeka/go2xunit

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
tests: fmt lint ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q $(GO) test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

test-xml: fmt lint | $(GO2XUNIT) ; $(info $(M) running xUnit tests…) @ ## Run tests with xUnit output
	$Q mkdir -p test
	$Q 2>&1 $(GO) test -timeout $(TIMEOUT)s -v $(TESTPKGS) | tee test/tests.output
	$(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

COVERAGE_MODE    = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML     = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML    = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint test-coverage-tools ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)
	$Q $(GO) test \
		-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $(TESTPKGS) | \
					grep '^$(MODULE)/' | \
					tr '\n' ',' | sed 's/,$$//') \
		-covermode=$(COVERAGE_MODE) \
		-coverprofile="$(COVERAGE_PROFILE)" $(TESTPKGS)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: | $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q $(GOLINT) -set_exit_status $(PKGS)

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$Q $(GO) fmt $(PKGS)

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(BIN)
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@echo "VARIABLES:"
	@echo "  APP = $(APP_NAME)"
	@echo "  OS = $(OS)"
	@echo "  ARCH = $(ARCH)"
	@echo "  REGISTRY = $(REGISTRY)"
	@echo
	@echo "TARGETS:"
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

# Docker

DOCKER_TARGETS := container publish publish-latest publish-version tag tag-version tag-latest release

.PHONY: $(DOCKER_TARGETS)
container: ## Build the container
	@docker build -t $(APP_NAME) .

publish: publish-latest publish-version ## Publish the `{version}` ans `latest` tagged containers to REGISTRY

publish-latest: tag-latest ## Publish the `latest` taged container to REGISTRY
	@echo 'publish latest to $(REGISTRY)'
	@docker push $(REGISTRY)/$(APP_NAME):latest

publish-version: tag-version ## Publish the `{version}` taged container to REGISTRY
	@echo 'publish $(VERSION) to $(REGISTRY)'
	@docker push $(REGISTRY)/$(APP_NAME):$(VERSION)

tag: tag-version tag-latest ## Generate container tags for the `{version}` ans `latest` tags

tag-latest: ## Generate container `{version}` tag
	@echo 'create tag latest'
	docker tag $(APP_NAME) $(REGISTRY)/$(APP_NAME):latest

tag-version: ## Generate container `latest` tag
	@echo 'create tag $(VERSION)'
	docker tag $(APP_NAME) $(REGISTRY)/$(APP_NAME):$(VERSION)

release: build publish ## Make a release by building and publishing the `{version}` ans `latest` tagged containers to REGISTRY