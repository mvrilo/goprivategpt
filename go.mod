module github.com/mvrilo/goprivategpt

go 1.20

require (
	github.com/go-skynet/go-llama.cpp v0.0.0-20230622210705-2c0a316c64f7
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/go-skynet/go-llama.cpp v0.0.0-20230622210705-2c0a316c64f7 => ./go-llama.cpp
