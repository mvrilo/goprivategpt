.PHONY: build clean

INCLUDE_PATH := $(abspath ./go-llama.cpp)
LIBRARY_PATH := $(abspath ./go-llama.cpp)

all: clean go-llama.cpp build-metal

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

clean:
	rm -rf go-llama.cpp goprivategpt
