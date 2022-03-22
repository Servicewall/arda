
all: build push
.PHONY: all build run test lint install-git-hook

build:
	@go build ./...

run:
	go run .

test:
	@go test -v ./...

lint:
	@if ! test -x "$$(command -v golangci-lint)"; then \
			echo "'golangci-lint' is not installed."; \
			echo "please install it first. go to https://golangci-lint.run/usage/install/"; \
			exit 2; \
		fi
	@golangci-lint run

install-git-hook:
	@if ! test -d ".git/hooks"; then \
			mkdir -p .git/hooks; \
		fi
	@curl -sLf -o .git/hooks/pre-commit "https://dev-res.s3.cn-northwest-1.amazonaws.com.cn/git-hooks/pre-commit"
	@chmod +x .git/hooks/pre-commit
	@echo "git hook is installed."

