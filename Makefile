export FLATNOTES_PATH=data
export FLATNOTES_USERNAME=user
export FLATNOTES_PASSWORD=pass
export FLATNOTES_SECRET_KEY=1

.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: watch
watch:
	# go install github.com/cespare/reflex@latest
	DEBUG=1 reflex --decoration=none --start-service -r '\.go$$' -- go run ./cmd/main.go

.PHONY: l2
l2:
	jsonnet --string --multi l2/ l2/l2.jsonnet

.PHONY: frontend
frontend:
	npm install
	npm run build


.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml ./...

.PHONY: test
test:
	go test -coverprofile=profile.out -coverpkg=github.com/rprtr258/flatnotes/internal/goldmark,github.com/rprtr258/flatnotes/internal/goldmark/ast,github.com/rprtr258/flatnotes/internal/goldmark/extension,github.com/rprtr258/flatnotes/internal/goldmark/extension/ast,github.com/rprtr258/flatnotes/internal/goldmark/parser,github.com/rprtr258/flatnotes/internal/goldmark/renderer,github.com/rprtr258/flatnotes/internal/goldmark/renderer/html,github.com/rprtr258/flatnotes/internal/goldmark/text,github.com/rprtr258/flatnotes/internal/goldmark/util ./...

cov: test
	go tool cover -html=profile.out

.PHONY: fuzz
fuzz:
	cd ./internal/goldmark/fuzz && go test -fuzz=Fuzz