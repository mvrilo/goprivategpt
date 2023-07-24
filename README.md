# goprivategpt

Implementation of PrivateGPT in Go using [langchaingo](https://github.com/tmc/langchaingo) and [go-llama.cpp](https://github.com/go-skynet/go-llama.cpp).

Supported file extensions:

- `txt`
- `csv`
- `html`

Supported vector stores:

- `weaviate`
- `sqlite-vss`

Requirements for running:

- ggml model

### Dependencies

brew install libomp

### Building

Mac:

`make build-metal`

Linux:

`make build`

### Usage

```
A way for you to interact with your documents.

Usage:
  goprivategpt [command]

Available Commands:
  ask         completes a given input
  help        Help about any command
  ingest      ingest documents from datastore
  server      starts the http server

Flags:
  -h, --help               help for goprivategpt
  -s, --storeaddr string   vector store address (default "localhost:8080")
  -t, --threads int        Number of threads (default 8)
  -n, --tokens int         Number of max tokens (default 512)

Use "goprivategpt [command] --help" for more information about a command.
```
