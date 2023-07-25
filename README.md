# goprivategpt

Implementation of PrivateGPT in Go using [langchaingo](https://github.com/tmc/langchaingo) and [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp).

Supported file extensions:

- `csv`
- `html`
- `txt`

Supported vector stores:

- [sqlite-vss](https://github.com/asg017/sqlite-vss)
- [weaviate](https://github.com/weaviate/weaviate)

Requirements for running:

- `ggml model`

### Dependencies

- `brew install libomp`

### Building

- Mac:

```
make build-metal
```

- Linux:

```
make build
```

### Usage

```
A way for you to interact with your documents.

Usage:
  goprivategpt [command]

Available Commands:
  ask         completes a given input
  help        Help about any command
  ingest      Ingests documents from source directory into the vector store
  server      Starts the http server

Flags:
  -h, --help               help for goprivategpt
  -s, --storeaddr string   Vector store filename or address (default "goprivategpt.db")
  -t, --threads int        Number of threads for LLM (default 8)
  -n, --tokens int         Number of max tokens in response (default 512)

Use "goprivategpt [command] --help" for more information about a command.
```
