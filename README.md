# goprivategpt

An implementation of privateGPT in Go.

### Usage

```
A way for you interact to your documents

Usage:
  goprivategpt [command]

Available Commands:
  ask         completes a given input
  help        Help about any command
  server      starts the http server

Flags:
  -h, --help           help for goprivategpt
  -m, --model string   Filepath of the model (default "models/GPT4All-13B-snoozy.ggmlv3.q4_0.bin")
  -t, --threads int    Number of threads (default 8)
  -n, --tokens int     Number of max tokens (default 512)

Use "goprivategpt [command] --help" for more information about a command.
```
