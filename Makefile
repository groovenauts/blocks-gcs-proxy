# See https://github.com/vincentbernat/hellogopher/blob/master/Makefile
GITHUB_ORG  = groovenauts
GITHUB_REPO = blocks-gcs-proxy
PACKAGE  = blocks-gcs-proxy
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= `grep VERSION version.go | cut -f2 -d\"`
GOPATH   = $(CURDIR)/.gopath~
BIN      = $(GOPATH)/bin
BASE     = $(GOPATH)/src/$(PACKAGE)
PKGS     = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "^$(PACKAGE)/vendor/"))
TESTPKGS = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
PKGDIR   = $(CURDIR)/pkg

export GOPATH

GO      = go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: fmt vendor | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && $(GO) build \
		-tags release \
		-o bin/$(PACKAGE) *.go

$(BASE): ; $(info $(M) setting GOPATH…)
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

# Tools

$(BIN):
	@mkdir -p $@
$(BIN)/%: $(BIN) | $(BASE) ; $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp go get $(REPOSITORY) && cp $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		rm -rf $$tmp ; exit $$ret

GODEP = $(BIN)/dep
$(BIN)/dep: REPOSITORY=github.com/golang/dep/cmd/dep

GOLINT = $(BIN)/golint
$(BIN)/golint: REPOSITORY=github.com/golang/lint/golint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: REPOSITORY=github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: REPOSITORY=github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: REPOSITORY=github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit: REPOSITORY=github.com/tebeka/go2xunit

GOX = $(BIN)/gox
$(BIN)/gox: REPOSITORY=github.com/mitchellh/gox

GHR = $(BIN)/ghr
$(BIN)/ghr: REPOSITORY=github.com/tcnksm/ghr


.PHONY: packages packages-tools
packages-tools: | $(GOX)
packages: OS_LIST := linux darwin
packages: ARCH := amd64
packages: PARALLEL := 2
packages: fmt vendor packages-tools | $(BASE) ; $(info $(M) Building packages…) @ ## Run gox
	mkdir -p ${PKGDIR} && \
	cd $(BASE) && \
		$(GOX) -output="${PKGDIR}/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="${OS_LIST}" -arch="${ARCH}" -parallel=${PARALLEL}

.PHONY: prerelease release release-tools
release-tools: | $(GHR)
release: git_guard packages release-tools | $(BASE) ; $(info $(M) Making a Release on github…) @ ## Run ghr
	$(GHR) -u $(GITHUB_ORG) -r $(GITHUB_REPO) --replace --draft ${VERSION} pkg
prerelease: git_guard packages release-tools | $(BASE) ; $(info $(M) Making a Pre-Release on github…) @ ## Run ghr
	$(GHR) -u $(GITHUB_ORG) -r $(GITHUB_REPO) --replace --draft --prerelease ${VERSION} pkg

# Tests

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt vendor | $(BASE) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(BASE) && $(GO) test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

test-xml: fmt vendor | $(BASE) $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	$Q cd $(BASE) && 2>&1 $(GO) test -timeout 20s -v $(TESTPKGS) | tee test/tests.output
	$(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt vendor test-coverage-tools | $(BASE) ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE) && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

vendor: Gopkg.toml Gopkg.lock | $(BASE) $(GODEP) ; $(info $(M) retrieving dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure
	@ln -nsf . vendor/src
	@touch $@
.PHONY: vendor-init
vendor-init: $(BASE)
	cd $(BASE) && dep init
.PHONY: vendor-update
vendor-update: vendor | $(BASE) $(GODEP)
ifeq "$(origin PKG)" "command line"
	$(info $(M) updating $(PKG) dependency…)
	$Q cd $(BASE) && $(GODEP) ensure -update $(PKG)
else
	$(info $(M) updating all dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure -update
endif
	@ln -nsf . vendor/src
	@touch vendor

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH)
	@rm -rf bin
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: git_guard git_tag git_push_tag tag
git_guard:
	$Q git diff --exit-code
git_tag:
	git tag v${VERSION}
git_push_tag:
	git push origin v${VERSION}
tag: git_tag git_push_tag


.PHONY: ci
ci:	fmt git_guard test
