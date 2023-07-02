.PHONY: all build build-metal clean docker lint full

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

full:
	make lint clean build-metal && \
	docker compose ps -aq | xargs -o docker rm -f; \
	sleep 2; \
	rm -rf ./data/weaviate/*; \
	docker compose up -d weaviate && \
	sleep 2; \
	echo 'ingesting --'; \
	./goprivategpt ingest && \
	sleep 2; \
	echo 'asking --'; \
	./goprivategpt ask -m ./models/wizardLM-7B.ggmlv3.q4_0.bin -p 'where murilo santana lives?'

clean:
	rm -rf go-llama.cpp goprivategpt
