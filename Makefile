.PHONY: build clean

INCLUDE_PATH := $(abspath ./gpt4all/gpt4all-bindings/golang)
LIBRARY_PATH := $(abspath ./gpt4all/gpt4all-bindings/golang)

all: clean gpt4all build

gpt4all: gpt4all/gpt4all-bindings/golang/libgpt4all.a

build:
	C_INCLUDE_PATH=$(INCLUDE_PATH) LIBRARY_PATH=$(LIBRARY_PATH) \
								 go build -o goprivategpt ./cmd/goprivategpt/main.go

gpt4all/gpt4all-bindings/golang/libgpt4all.a:
	git clone --recursive https://github.com/nomic-ai/gpt4all
	(cd gpt4all/gpt4all-bindings/golang; make libgpt4all.a || true)

clean:
	rm -rf gpt4all
