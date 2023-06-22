module github.com/mvrilo/goprivategpt

go 1.20

require (
	github.com/nomic-ai/gpt4all/gpt4all-bindings/golang v0.0.0-20230622161949-a4230754039f
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/nomic-ai/gpt4all/gpt4all-bindings/golang v0.0.0-20230622161949-a4230754039f => ./gpt4all/gpt4all-bindings/golang
