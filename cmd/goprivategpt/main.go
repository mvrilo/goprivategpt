package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"

	goprivategpt "github.com/mvrilo/goprivategpt/privategpt"
	"github.com/spf13/cobra"
)

var (
	threads int
	tokens  int
	model   string
	prompt  string
	addr    string
)

func privategpt(model string, threads, tokens int) *goprivategpt.PrivateGPT {
	pgpt, err := goprivategpt.New(model, threads, tokens)
	if err != nil {
		log.Fatal(err)
	}
	return pgpt
}

func main() {
	rootCmd := &cobra.Command{
		Use:                   "goprivategpt",
		Short:                 "A way for you interact to your documents",
		DisableAutoGenTag:     true,
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&model, "model", "m", "models/GPT4All-13B-snoozy.ggmlv3.q4_0.bin", "Filepath of the model")
	flags.IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads")
	flags.IntVarP(&tokens, "tokens", "n", 512, "Number of max tokens")

	askCmd := &cobra.Command{
		Use:   "ask",
		Short: "completes a given input",
		Run: func(cmd *cobra.Command, args []string) {
			pgpt := privategpt(model, threads, tokens)

			text := cmd.Flag("prompt").Value.String()
			err := pgpt.Predict(text)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(pgpt.Response())
		},
	}
	askCmd.PersistentFlags().StringVarP(&prompt, "prompt", "p", "", "input text")

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "starts the http server",
		Run: func(cmd *cobra.Command, args []string) {
			pgpt := privategpt(model, threads, tokens)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			go func() {
				if err := pgpt.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal(err)
				}
			}()

			select {
			case <-ctx.Done():
				pgpt.Shutdown()
			}
		},
	}
	serverCmd.PersistentFlags().StringVarP(&addr, "address", "a", ":8000", "address of the http server")

	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
