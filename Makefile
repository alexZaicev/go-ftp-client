CMD_DIR         := "./cmd"
DIST_DIR        := "./dist"
INTERNAL_DIR    := "./internal"
MOCKS_DIR       := "./mocks"

INTERNAL_NON_TEST_GO_FILES = $(shell find $(INTERNAL_DIR) -type f -name '*.go' -not -name '*_test.go')

MOCKERY      := mockery
MOCKERY_ARGS := --all --keeptree --dir $(INTERNAL_DIR)

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: checks
checks: mocks fmt # wire

.PHONY: build
build: vendor
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR) $(CMD_DIR)/...

.PHONY: unit
unit: vendor
	go test $(INTERNAL_DIR)/... $(CMD_DIR)/... \
  		-cover \
    	-coverprofile=coverage.out \
    	-count=1
	@cat coverage.out | \
		awk 'BEGIN {cov=0; stat=0;} $$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } \
    	END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}'
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt: vendor
	gofmt -s -w -e $(CMD_DIR) $(FUNCTIONAL_TESTS_DIR) $(INTERNAL_DIR) $(CONTRACT_TEST_DIR)
	gci write \
        -s Standard \
        -s Default \
        -s 'Prefix(github.com)' \
        -s 'Prefix(github.com/alexZaicev/go-ftp-client)' \
        $(CMD_DIR) $(INTERNAL_DIR)
	goimports -local github.hpe.com -w $(CMD_DIR) $(INTERNAL_DIR)

.PHONY: golint
golint:
	golangci-lint run --timeout=3m --max-issues-per-linter 0 --max-same-issues 0

.PHONY: mocks
mocks: $(INTERNAL_NON_TEST_GO_FILES)
	rm -rf $(MOCKS_DIR)_maketemp/
	@# Mockery returns error code 0 on these errors but produces incorrect output
	if $(MOCKERY) $(MOCKERY_ARGS) --output $(MOCKS_DIR)_maketemp 2>&1 | grep ERR; then \
		rm -rf $(MOCKS_DIR)_maketemp; \
		exit 1; \
	fi
	rm -rf $(MOCKS_DIR)/
	mv $(MOCKS_DIR)_maketemp $(MOCKS_DIR)