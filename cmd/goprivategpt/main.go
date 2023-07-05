package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	gollama "github.com/mvrilo/goprivategpt/langchaingo-gollamacpp"
	goprivategpt "github.com/mvrilo/goprivategpt/privategpt"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

var (
	threads    int
	tokens     int
	model      string
	prompt     string
	datadir    string
	storeaddr  string
	serveraddr string
)

func privategpt(withLLM bool) *goprivategpt.PrivateGPT {
	var (
		llm llms.LanguageModel
		err error
	)

	if withLLM {
		llm, err = gollama.NewLLM(model, threads, true)
		if err != nil {
			log.Fatal(err)
		}
	}

	emb, err := embeddings.NewOpenAI()
	if err != nil {
		log.Fatal(err)
	}
	store, err := weaviate.New(
		weaviate.WithScheme("http"),
		weaviate.WithHost(storeaddr),
		weaviate.WithEmbedder(emb),
		weaviate.WithIndexName("PGPT"),
		weaviate.WithTextKey("text"),
		weaviate.WithNameSpaceKey("docs"),
	)
	if err != nil {
		log.Fatal(err)
	}

	pgpt, err := goprivategpt.New(model, llm, store)
	if err != nil {
		log.Fatal(err)
	}
	return pgpt
}

func main() {
	rootCmd := &cobra.Command{
		Use:                   "goprivategpt",
		Short:                 "A way for you to interact with your documents.",
		DisableAutoGenTag:     true,
		DisableSuggestions:    true,
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&storeaddr, "storeaddr", "s", "localhost:8080", "Vector store address")
	flags.IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads for LLM")
	flags.IntVarP(&tokens, "tokens", "n", 512, "Number of max tokens in response")

	askCmd := &cobra.Command{
		Use:   "ask",
		Short: "completes a given input",
		Run: func(cmd *cobra.Command, args []string) {
			input := cmd.Flag("prompt").Value.String()
			if input == "" {
				cmd.Help()
				return
			}
			pgpt := privategpt(true)
			res, err := pgpt.Predict(context.Background(), input)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Result:", res)
			os.Exit(0)
		},
	}
	askFlags := askCmd.PersistentFlags()
	askFlags.StringVarP(&prompt, "prompt", "p", "", "input text")
	askFlags.StringVarP(&model, "model", "m", "models/open-llama-7B-open-instruct.ggmlv3.q2_K.bin", "Filepath of the model")

	ingestCmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingests documents from source directory into the vector store",
		Run: func(cmd *cobra.Command, args []string) {
			datadir := cmd.Flag("source_dir").Value.String()
			if datadir == "" {
				cmd.Help()
				return
			}
			pgpt := privategpt(false)
			err := pgpt.IngestDocuments(context.Background(), datadir)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		},
	}
	ingestCmd.PersistentFlags().StringVarP(&datadir, "source_dir", "i", "./documents", "Directory to ingest documents")

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Starts the http server",
		Run: func(cmd *cobra.Command, args []string) {
			pgpt := privategpt(true)
			server, err := goprivategpt.NewServer(pgpt)
			if err != nil {
				log.Fatal(err)
			}
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			go func() {
				if err := server.Start(serveraddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal(err)
				}
			}()

			select {
			case <-ctx.Done():
				server.Shutdown()
			}
		},
	}
	serverFlags := serverCmd.PersistentFlags()
	serverFlags.StringVarP(&serveraddr, "address", "a", ":8000", "address of the http server")
	serverFlags.StringVarP(&model, "model", "m", "models/open-llama-7B-open-instruct.ggmlv3.q2_K.bin", "Filepath of the model")

	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
