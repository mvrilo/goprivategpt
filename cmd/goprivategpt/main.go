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
	"strings"
	"syscall"

	"github.com/mvrilo/goprivategpt/llama"
	goprivategpt "github.com/mvrilo/goprivategpt/privategpt"
	"github.com/mvrilo/goprivategpt/sqlitevss"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/weaviate"
)

// const defaultModel = "models/orca-mini-7b.ggmlv3.q4_0.bin"
const defaultModel = "models/orca-mini-v2_7b.ggmlv3.q5_1.bin"

// const defaultModel = "models/llama-2-7b.ggmlv3.q4_K_S.bin"

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
	llm, err := llama.NewLLM(model, threads, 1024, true)
	if err != nil {
		log.Fatal(err)
	}

	var store vectorstores.VectorStore
	if storeaddr == ":memory:" || !strings.HasPrefix(storeaddr, "http") {
		store, err = sqlitevss.New(storeaddr)
	} else {
		emb, err := embeddings.NewOpenAI()
		if err != nil {
			log.Fatal(err)
		}

		store, err = weaviate.New(
			weaviate.WithScheme("http"),
			weaviate.WithHost(storeaddr),
			weaviate.WithEmbedder(emb),
			weaviate.WithIndexName("PGPT"),
			weaviate.WithTextKey("text"),
			weaviate.WithNameSpaceKey("docs"),
		)
	}

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
	flags.StringVarP(&storeaddr, "storeaddr", "s", "goprivategpt.db", "Vector store address")
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
			fmt.Println(res)
			os.Exit(0)
		},
	}
	askFlags := askCmd.PersistentFlags()
	askFlags.StringVarP(&prompt, "prompt", "p", "", "input text")
	askFlags.StringVarP(&model, "model", "m", defaultModel, "Filepath of the model")

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
			// if st, ok := pgpt.Store.(interface{ Close() error }); ok {
			// 	defer st.Close()
			// }
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
			if st, ok := pgpt.Store.(interface{ Close() error }); ok {
				defer st.Close()
			}

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
	serverFlags.StringVarP(&model, "model", "m", defaultModel, "Filepath of the model")

	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
