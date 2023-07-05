package gollamacpp

import (
	"bytes"
	"context"
	"errors"

	gollama "github.com/go-skynet/go-llama.cpp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

type LLM struct {
	buf     *bytes.Buffer
	client  *gollama.LLama
	caching bool
	threads int
}

// New returns a new LLama LLM.
func NewLLM(model string, threads int, caching bool) (*LLM, error) {
	client, err := gollama.New(
		model,
		gollama.EnableF16Memory,
		gollama.SetContext(512),
		gollama.EnableEmbeddings,
		gollama.SetGPULayers(0),
	)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	llm := &LLM{
		client:  client,
		buf:     buf,
		threads: threads,
	}
	client.SetTokenCallback(llm.tokenCallback)
	return llm, nil
}

func (o *LLM) tokenCallback(token string) bool {
	_, err := o.buf.WriteString(token)
	if err != nil {
		print(token)
		// todo: log
		return false
	}
	return true
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

// Generate generates a completion for the given prompt.
func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	gollamaOpts := []gollama.PredictOption{
		gollama.SetTokens(opts.MaxTokens),
		gollama.SetStopWords(opts.StopWords...),
		gollama.SetThreads(o.threads),
	}

	if o.caching {
		gollamaOpts = append(gollamaOpts, gollama.EnablePromptCacheAll)
	}

	result, err := o.client.Predict(
		prompts[0],
		gollamaOpts...,
	)
	if err != nil {
		return nil, err
	}
	return []*llms.Generation{
		{Text: result},
	}, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, prompts []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, prompts, options...)
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens("gpt2", text)
}
