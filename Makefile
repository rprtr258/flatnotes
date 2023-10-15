export FLATNOTES_PATH=data
export FLATNOTES_USERNAME=user
export FLATNOTES_PASSWORD=pass
export FLATNOTES_SECRET_KEY=1

.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: watch
watch:
	reflex --start-service -r '\.go$$' -- go run ./cmd/main.go

.PHONY: l2
l2:
	jsonnet --string --multi l2/ l2/l2.jsonnet
