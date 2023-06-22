package main

import (
	"fmt"
	"log"
	"runtime"

	goprivategpt "github.com/mvrilo/goprivategpt/privategpt"
	"github.com/spf13/cobra"
)

var (
	model   string
	threads int
	tokens  int
	prompt  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "goprivategpt",
		Short: "goprivategpt is a way for you interact to your documents",
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&model, "model", "m", "models/GPT4All-13B-snoozy.ggmlv3.q4_0.bin", "Filepath of the model")
	flags.IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads")
	flags.IntVarP(&tokens, "tokens", "n", 512, "Number of max tokens")

	pgpt, err := goprivategpt.New(model, threads, tokens)
	if err != nil {
		log.Fatal(err)
	}

	err = pgpt.Load()
	if err != nil {
		log.Fatal(err)
	}

	askCmd := &cobra.Command{
		Use:   "ask",
		Short: "ask completes a given input",
		Run: func(cmd *cobra.Command, args []string) {
			text := cmd.Flag("prompt").Value.String()
			err := pgpt.Predict(text)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(pgpt.Out.String())
		},
	}
	askCmd.PersistentFlags().StringVarP(&prompt, "prompt", "p", "", "input text")
	rootCmd.AddCommand(askCmd)

	if err = rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
