# ajfs Makefile

# Directory in which the temporary files are stored
BUILD_OUTPUT_DIR=build
EXEC_OUTPUT_DIR=${BUILD_OUTPUT_DIR}/bin
REPORT_OUTPUT_DIR=${BUILD_OUTPUT_DIR}/reports
# Go module name for this repo
MODULE_NAME="github.com/andrejacobs/ajfs"
# cmd/{APP}/main.go
INPUT_SRC_FILE="main.go"

# Clean, Build and Test
.PHONY: all
all: clean build test

# Remove build output
.PHONY: clean
clean:
	@echo "Cleaning build output"
	go clean
	rm -rf ${BUILD_OUTPUT_DIR}

# Gather info about the current platform and version for the app
.PHONY: versioninfo
versioninfo:
	$(eval CURRENT_OS := $(shell uname -s))
	$(eval CURRENT_CPU_ARCH := $(shell uname -p))

	$(eval GIT_COMMIT_HASH := $(shell git rev-parse HEAD))
	$(eval GIT_TAG := $(shell git describe --tags --dirty --always))

	$(eval GO_LDFLAGS := -s -w -X ${MODULE_NAME}/internal/compiledinfo.Version=${GIT_TAG} -X ${MODULE_NAME}/internal/compiledinfo.GitCommitHash=${GIT_COMMIT_HASH})

# Display info about the current platform and configurations etc.
.PHONY: info
info: versioninfo
	@echo "Information"
	@echo "OS: ${CURRENT_OS}"
	@echo "CPU architecture: ${CURRENT_CPU_ARCH}"
	@echo "GIT_COMMIT_HASH: ${GIT_COMMIT_HASH}"
	@echo "GIT_TAG: ${GIT_TAG}"
	@echo "GO_LDFLAGS: ${GO_LDFLAGS}"

# Build executables for the current platform
.PHONY: build
build: versioninfo
	@echo "Building for the current platform: ${CURRENT_OS}/${CURRENT_CPU_ARCH}"
	$(eval CURRENT_OUTPUT_DIR := ${EXEC_OUTPUT_DIR}/${CURRENT_OS}/${CURRENT_CPU_ARCH})

# Each directory inside of ./cmd is considered to be a seperate CLI executable
	@for name in cmd/*; do \
		CURRENT_EXECUTABLE=${CURRENT_OUTPUT_DIR}/$${name#cmd/}; \
		GO_LDFLAGS="${GO_LDFLAGS} -X ${MODULE_NAME}/internal/compiledinfo.AppName=$${name#cmd/}"; \
    	echo "Compiling $${name} to: $${CURRENT_EXECUTABLE}"; \
		mkdir -p ${CURRENT_OUTPUT_DIR}; \
		CGO_ENABLED=0 go build -ldflags "$${GO_LDFLAGS}" -o $${CURRENT_EXECUTABLE} "$${name}/${INPUT_SRC_FILE}"; \
		echo "Linking $${CURRENT_EXECUTABLE} to: ${EXEC_OUTPUT_DIR}/$${name#cmd/}"; \
		ABS_CURRENT_EXECUTABLE=`realpath $${CURRENT_EXECUTABLE}`; \
		ln -sf "$${ABS_CURRENT_EXECUTABLE}" "${EXEC_OUTPUT_DIR}/$${name#cmd/}"; \
	done

# Install each executable in the go/bin directory
.PHONY: install
install: build
	$(eval GO_BIN_DIR := $(shell go env GOPATH)/bin)
	mkdir -p "${GO_BIN_DIR}"
	@for name in ${CURRENT_OUTPUT_DIR}/*; do \
		echo "Copying $${name} to ${GO_BIN_DIR}/"; \
		cp "$${name}" "${GO_BIN_DIR}/"; \
	done

# Generate documentation
.PHONY: docs
docs:
	@rm -rf ./docs/cli
	@go run ./internal/docgen -out ./docs/cli/md -format markdown

#------------------------------------------------------------------------------
# Code quality assurance
#------------------------------------------------------------------------------

# Run unit-testing with race detector and code coverage report
.PHONY: test
test:
	@echo "Running unit-tests"
	$(eval COVERAGE_REPORT := ${REPORT_OUTPUT_DIR}/codecoverage)
	@mkdir -p "${REPORT_OUTPUT_DIR}"
	@go test -v -count=1 -race ./... -coverprofile="${COVERAGE_REPORT}"

# Check that the source code is formatted correctly according to the gofmt standards
.PHONY: check-formatting
check-formatting:
	@test -z $(shell gofmt -e -l ./ | tee /dev/stderr) || (echo "Please fix formatting first with gofmt" && exit 1)

# Check for other possible issues in the code
# NOTE: To install golangci-lint
# https://golangci-lint.run/welcome/install/#local-installation
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5
.PHONY: check-lint
check-lint:
	@echo "Linting code"
	go vet ./...
ifneq (${CI}, true)
	golangci-lint run
	addlicense -check -c "Andre Jacobs" -l mit -ignore '.github/**' -ignore 'build/**' ./
endif

# Check code quality
.PHONY: check
check: check-formatting check-lint

#------------------------------------------------------------------------------
# Go miscellaneous
#------------------------------------------------------------------------------

# Fetch required go modules
.PHONY: go-deps
go-deps:
	go mod download

# Tidy up module references (also donwloads deps)
.PHONY: go-tidy
go-tidy:
	go mod tidy

# Add the copyright and license notice
.PHONY: addlic
addlic:
	@echo "Adding copyright and license notice"
	addlicense -v -c "Andre Jacobs" -l mit -ignore '.github/**' -ignore 'build/**' ./
