.PHONY: build
build:
	go build -o build/hedgehog cmd/hedgehog/main.go

.PHONY: vhs
vhs: build
	vhs .vhs.tape
