BINARY := bin/skycast
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run test lint demo gif clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run . $(ARGS)

test:
	go test ./... -race

lint:
	golangci-lint run

demo:
	go run . --demo

# Record with VHS, then shrink: a 48-color palette and frame diffs keep
# the GIF under 5 MB without visibly changing the scenes.
gif:
	vhs demo.tape
	ffmpeg -v error -y -i assets/demo.gif -filter_complex \
		"[0:v]split[a][b];[a]palettegen=max_colors=48:stats_mode=diff[p];[b][p]paletteuse=dither=none:diff_mode=rectangle" \
		-gifflags +transdiff assets/demo.min.gif
	mv assets/demo.min.gif assets/demo.gif

clean:
	rm -rf bin dist
