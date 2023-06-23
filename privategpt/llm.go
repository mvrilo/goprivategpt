package goprivategpt

import (
	"strings"

	gollama "github.com/go-skynet/go-llama.cpp"
)

type LLMConfig struct {
	Model   string
	Threads int
	Tokens  int
	TopK    int
	TopP    float64
}

type LLM struct {
	config *LLMConfig
	client *gollama.LLama
	Out    strings.Builder
}

func NewLLM(model string, threads, tokens int) (*LLM, error) {
	conf := LLMConfig{
		Model:   model,
		Threads: threads,
		Tokens:  tokens,
		TopK:    90,
		TopP:    0.86,
	}

	cli, err := gollama.New(
		model,
		gollama.EnableF16Memory,
		gollama.SetContext(128),
		gollama.EnableEmbeddings,
		gollama.SetGPULayers(0),
	)
	if err != nil {
		return nil, err
	}

	llm := &LLM{
		config: &conf,
		client: cli,
	}

	cli.SetTokenCallback(llm.tokenCallback)
	return llm, nil
}

func (l *LLM) tokenCallback(token string) bool {
	l.Out.WriteString(token)
	return true
}

func (l *LLM) Response() string {
	return l.Out.String()
}

func (l *LLM) Predict(input string) error {
	_, err := l.client.Predict(
		input,
		gollama.Debug,
		gollama.SetThreads(l.config.Threads),
		gollama.SetTokens(l.config.Tokens),
		gollama.SetTopK(l.config.TopK),
		gollama.SetTopP(l.config.TopP),
	)
	return err
}
