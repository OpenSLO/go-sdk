.DEFAULT_GOAL := help
MAKEFLAGS += --silent --no-print-directory

BIN_DIR := ./.bin
SCRIPTS_DIR := ./internal/scripts

# Print Makefile target step description for check.
# Only print 'check' steps this way, and not dependent steps, like 'install'.
# ${1} - step description
define _print_step
	printf -- '------\n%s...\n' "${1}"
endef

## Activate developer environment using devbox. Run `make install/devbox` first If you don't have devbox installed.
activate:
	devbox shell

## Install devbox binary.
install/devbox:
	curl -fsSL https://get.jetpack.io/devbox | bash

## Automatically load devbox environment, requires direnv.
install/direnv:
	devbox generate direnv

.PHONY: test test/go/unit
## Run all tests.
test: test/go/unit

## Run Go unit tests.
test/go/unit:
	$(call _print_step,Running Go unit tests)
	go test -race -cover ./...

.PHONY: check check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate
## Run all checks.
check: check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate

## Run 'go vet' on the whole project.
check/vet:
	$(call _print_step,Running go vet)
	go vet ./...

## Run golangci-lint all-in-one linter with configuration defined inside .golangci.yml.
check/lint:
	$(call _print_step,Running golangci-lint)
	golangci-lint run

## Check for security problems using gosec, which inspects the Go code by scanning the AST.
check/gosec:
	$(call _print_step,Running gosec)
	gosec -exclude-dir=test -exclude-generated -quiet ./...

## Check spelling, rules are defined in cspell.json.
check/spell:
	$(call _print_step,Verifying spelling)
	yarn --silent cspell --no-progress '**/**'

## Check for trailing whitespaces in any of the projects' files.
check/trailing:
	$(call _print_step,Looking for trailing whitespaces)
	$(SCRIPTS_DIR)/check-trailing-whitespaces.bash

## Check markdown files for potential issues with markdownlint.
check/markdown:
	$(call _print_step,Verifying Markdown files)
	yarn --silent markdownlint '**/*.md' --ignore 'node_modules'

## Check for potential vulnerabilities across all Go dependencies.
check/vulns:
	$(call _print_step,Running govulncheck)
	govulncheck ./...

## Verify if the files are formatted.
## You must first commit the changes, otherwise it won't detect the diffs.
check/format:
	$(call _print_step,Checking if files are formatted)
	$(SCRIPTS_DIR)/check-formatting.sh

## Verify if the auto generated code has been committed.
check/generate:
	$(call _print_step,Checking if generated code matches the provided definitions)
	$(SCRIPTS_DIR)/check-generate.sh

.PHONY: generate generate/go generate/govydoc
## Auto generate files.
generate: generate/go generate/govydoc

## Generate Golang code.
generate/go:
	$(call _print_step,Generating Go code)
	go generate ./...

## Generate object docs using govydoc.
generate/govydoc:
	$(call _print_step,Generating object docs)
	go run ./internal/cmd/objectdoc/main.go > ./internal/cmd/objectdoc/docs.json

.PHONY: format format/go format/cspell
## Format files.
format: format/go format/cspell

## Format Go files.
format/go:
	$(call _print_step,Formatting Go files)
	gofumpt -l -w -extra .
	goimports -local=$$(head -1 go.mod | awk '{print $$2}') -w .
	golines -m 120 --ignore-generated --reformat-tags -w .

## Format cspell config file.
format/cspell:
	$(call _print_step,Formatting cspell.yaml configuration \(words list\))
	yarn --silent format-cspell-config

.PHONY: install
## Install all dev dependencies.
install: install/yarn

## Install JS dependencies with yarn.
install/yarn:
	$(call _print_step,Installing yarn dependencies)
	yarn --silent install
	
.PHONY: help
## Print this help message.
help:
	$(SCRIPTS_DIR)/makefile-help.awk $(MAKEFILE_LIST)
