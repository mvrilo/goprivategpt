.PHONY: all build build-metal clean docker lint fullcheck cdeps

SQLITE_VSS := $(abspath ./sqlite-vss/dist/release)
LLAMA := $(abspath ./go-llama.cpp)

INCLUDE_PATH := $(abspath ./go-llama.cpp)
LIBRARY_PATH := $(abspath ./go-llama.cpp)
BUILD_FLAGS := C_INCLUDE_PATH=${LLAMA}:${SQLITE_VSS} LIBRARY_PATH=${LLAMA}:${SQLITE_VSS}
METAL_FLAGS := -framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders
CGO_FLAGS := -L$(abspath ./go-llama.cpp) -L$(abspath ./sqlite-vss/dist/release) -L/opt/homebrew/opt/libomp/lib
CGO_METAL_FLAGS := ${CGO_FLAGS} ${METAL_FLAGS}

all: build-metal

lint:
	go vet ./...

build-metal: cdeps
	(cd go-llama.cpp || exit 1; BUILD_TYPE=metal LLAMA_METAL=1 make clean libbinding.a);
	cp go-llama.cpp/llama.cpp/ggml-metal.metal ./ggml-metal.metal;
	CGO_LDFLAGS='${CGO_METAL_FLAGS}' ${BUILD_FLAGS} \
		go build -o goprivategpt ./cmd/goprivategpt/main.go

build: cdeps
	(cd go-llama.cpp || exit 1; make clean libbinding.a);
	CGO_LDFLAGS='${CGO_FLAGS}' ${BUILD_FLAGS} \
		go build -o goprivategpt ./cmd/goprivategpt/main.go

cdeps: sqlite-vss go-llama.cpp

sqlite-vss:
	git clone --recursive https://github.com/asg017/sqlite-vss
	(cd sqlite-vss; /bin/sh ./vendor/get_sqlite.sh) && \
		(cd sqlite-vss/vendor/sqlite; ./configure && make) && \
		(cd sqlite-vss; make loadable-release)

go-llama.cpp:
	git clone --recursive https://github.com/go-skynet/go-llama.cpp

# docker:
# 	docker build . -t mvrilo/goprivategpt

fullcheck:
	@( \
		echo 'Build goprivategpt'; \
		make lint clean build-metal >/dev/null 2>/dev/null || exit 1; \
		echo 'Cleaning up weaviate container'; \
		docker compose -f ./testdata/docker-compose.yml ps weaviate -q 2>/dev/null | xargs -o docker rm -f 2>/dev/null >/dev/null ; \
		rm -rf ./testdata/tmp/* || true 2>/dev/null; \
		mkdir ./testdata/tmp/goprivategpt_weaviate_test || true 2>/dev/null; \
		echo 'Deploying weaviate container'; \
		docker compose -f ./testdata/docker-compose.yml up -d weaviate >/dev/null && \
		sleep 10; \
		echo 'Ingesting documents from ./testdata' || exit 1; \
		./goprivategpt ingest -i ./testdata/docs 2>/dev/null >/dev/null && \
		sleep 5; \
		echo 'Prompt: What damage did zero cool cause?'; \
		time ./goprivategpt ask -p 'What damage did zero cool cause?' \
		)

clean:
	rm -rf ./go-llama.cpp ./sqlite-vss ./goprivategpt
