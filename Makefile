.PHONY: all build build-metal clean docker lint fullcheck

INCLUDE_PATH := $(abspath ./go-llama.cpp)
LIBRARY_PATH := $(abspath ./go-llama.cpp)

all: lint clean build-metal

lint:
	go vet ./...

build-metal: go-llama.cpp
	(cd go-llama.cpp || exit 1; BUILD_TYPE=metal make clean libbinding.a);
	cp go-llama.cpp/llama.cpp/ggml-metal.metal ./ggml-metal.metal;
	C_INCLUDE_PATH=${INCLUDE_PATH} CGO_LDFLAGS=${CGO_LDFLAGS} LIBRARY_PATH=${LIBRARY_PATH} \
								 CGO_LDFLAGS='-framework Foundation -framework Metal -framework MetalKit -framework MetalPerformanceShaders' \
								 go build -o goprivategpt ./cmd/goprivategpt/main.go

build: go-llama.cpp
	(cd go-llama.cpp || exit 1; make clean libbinding.a);
	C_INCLUDE_PATH=${INCLUDE_PATH} CGO_LDFLAGS=${CGO_LDFLAGS} LIBRARY_PATH=${LIBRARY_PATH} \
								 go build -o goprivategpt ./cmd/goprivategpt/main.go

go-llama.cpp:
	git clone --recursive https://github.com/go-skynet/go-llama.cpp

docker:
	docker build . -t mvrilo/goprivategpt

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
	rm -rf ./go-llama.cpp ./goprivategpt
