export FLATNOTES_PATH=data
export FLATNOTES_USERNAME=user
export FLATNOTES_PASSWORD=pass
export FLATNOTES_SECRET_KEY=1

.PHONY: run
run:
	make frontend
	mkdir -p data
	go run ./cmd/main.go

.PHONY: watch
watch:
	DEBUG=1 go tool reflex --decoration=none --start-service -v -r '\.(html|go|vue|css|scss|ts|Makefile)$$' -R '/dist/' -R 'node_modules/' -- make run

.PHONY: frontend
frontend:
	PATH="/home/linuxbrew/.linuxbrew/lib/node_modules/npm/bin/node-gyp-bin:$$PATH" bun i
	bun run build

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: format
format:
	golangci-lint fmt -c .golangci.yml ./...

.PHONY: test
test:
	go test -coverprofile=profile.out -coverpkg=github.com/rprtr258/flatnotes/internal/goldmark,github.com/rprtr258/flatnotes/internal/goldmark/ast,github.com/rprtr258/flatnotes/internal/goldmark/extension,github.com/rprtr258/flatnotes/internal/goldmark/extension/ast,github.com/rprtr258/flatnotes/internal/goldmark/parser,github.com/rprtr258/flatnotes/internal/goldmark/renderer,github.com/rprtr258/flatnotes/internal/goldmark/renderer/html,github.com/rprtr258/flatnotes/internal/goldmark/text,github.com/rprtr258/flatnotes/internal/goldmark/util ./...

cov: test
	go tool cover -html=profile.out
