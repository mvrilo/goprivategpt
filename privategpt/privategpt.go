package goprivategpt

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

func newLoader(ext string, r io.Reader) documentloaders.Loader {
	if len(ext) < 2 {
		return nil
	}

	switch ext[1:] {
	case "pdf":
		rr, ok := r.(io.ReaderAt)
		if !ok {
			return nil
		}
		return documentloaders.NewPDF(rr, 0)
	case "md", "markdown", "html", "htm":
		return documentloaders.NewHTML(r)
	case "text", "txt":
		return documentloaders.NewText(r)
	case "csv":
		return documentloaders.NewCSV(r)
	default:
		return nil
	}
}

type PrivateGPT struct {
	LLM   llms.LanguageModel
	Store vectorstores.VectorStore
	Model string
}

func New(model string, llm llms.LanguageModel, store vectorstores.VectorStore) (*PrivateGPT, error) {
	return &PrivateGPT{
		LLM:   llm,
		Store: store,
		Model: model,
	}, nil
}

func (p *PrivateGPT) IngestDocuments(ctx context.Context, datadir string) error {
	docs, err := p.LoadDocuments(ctx, datadir)
	if err != nil {
		return err
	}
	return p.Store.AddDocuments(ctx, docs, vectorstores.WithNameSpace("docs"))
}

func (p *PrivateGPT) LoadDocuments(ctx context.Context, datadir string) ([]schema.Document, error) {
	var docs []schema.Document
	err := fs.WalkDir(os.DirFS(datadir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dir, _ := filepath.Abs(datadir)
		full := dir + "/" + path
		f, err := os.Open(full)
		if err != nil {
			return err
		}
		defer f.Close()

		ext := filepath.Ext(full)
		loader := newLoader(ext, f)
		if loader == nil {
			// skip
			// fmt.Printf("no loader found for ext: %s\n", ext)
			return nil
		}

		doc, err := loader.Load(ctx)
		if err != nil {
			return err
		}

		docs = append(docs, doc...)
		return nil
	})
	return docs, err
}

func (p *PrivateGPT) Predict(ctx context.Context, input string) (string, error) {
	template := `You are a virtual Assistant that responds to questions based on some context extracted from documents that the User provided.
If you know the answer, be direct. If you don't now the answer, just reply something like: "I don't know".


## Context:
{{.context}}


## User input:
{{.question}}


## Assistant answer:`

	prompt := prompts.NewPromptTemplate(template, []string{"question", "context"})
	combineChain := chains.NewStuffDocuments(chains.NewLLMChain(p.LLM, prompt))

	retriever := vectorstores.ToRetriever(p.Store, 20, vectorstores.WithNameSpace("docs"))
	retrievalChain := chains.NewRetrievalQA(combineChain, retriever)

	result, err := chains.Run(ctx, retrievalChain, input, chains.WithModel(p.Model))
	if err != nil {
		return "", err
	}

	return result, nil
}
