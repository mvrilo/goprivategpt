.PHONY: build clean

INCLUDE_PATH := $(abspath ./go-llama.cpp)
LIBRARY_PATH := $(abspath ./go-llama.cpp)

all: clean go-llama.cpp build

build-metal: go-llama.cpp
	(cd go-llama.cpp || exit 1; BUILD_TYPE=metal make clean libbinding.a);
	cp go-llama.cpp/llama.cpp/ggml-metal.metal ./ggml-metal.metal;
	C_INCLUDE_PATH=${INCLUDE_PATH} CGO_LDFLAGS=${CGO_LDFLAGS} LIBRARY_PATH=${LIBRARY_PATH} \
								 CGO_LDFLAGS='-framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders' \
								 go build -o goprivategpt ./cmd/goprivategpt/main.go

build:
	C_INCLUDE_PATH=${INCLUDE_PATH} CGO_LDFLAGS=${CGO_LDFLAGS} LIBRARY_PATH=${LIBRARY_PATH} \
								 go build -o goprivategpt ./cmd/goprivategpt/main.go

go-llama.cpp:
	git clone --recursive https://github.com/go-skynet/go-llama.cpp

gpt4all: gpt4all/gpt4all-bindings/golang/libgpt4all.a

gpt4all/gpt4all-bindings/golang/libgpt4all.a:
	git clone --recursive https://github.com/nomic-ai/gpt4all
	(cd gpt4all/gpt4all-bindings/golang; make libgpt4all.a || true)

clean:
	rm -rf gpt4all go-llama.cpp
