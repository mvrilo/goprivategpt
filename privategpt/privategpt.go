package goprivategpt

import (
	"strings"

	gollama "github.com/go-skynet/go-llama.cpp"
)

type PrivateGPT struct {
	Model   string
	Threads int
	Tokens  int
	TopK    int
	TopP    float64

	client *gollama.LLama
	Out    strings.Builder
}

func New(model string, threads, tokens int) (*PrivateGPT, error) {
	return &PrivateGPT{
		Model:   model,
		Threads: threads,
		Tokens:  tokens,
		TopK:    90,
		TopP:    0.86,
	}, nil
}

func (p *PrivateGPT) tokenCallback(token string) bool {
	p.Out.WriteString(token)
	return true
}

func (p *PrivateGPT) Load() error {
	l, err := gollama.New(p.Model)
	if err != nil {
		return err
	}
	l.SetTokenCallback(p.tokenCallback)
	p.client = l
	return nil
}

func (p *PrivateGPT) Predict(input string) error {
	_, err := p.client.Predict(
		input,
		gollama.SetThreads(p.Threads),
		gollama.SetTokens(p.Tokens),
		gollama.SetTopK(p.TopK),
		gollama.SetTopP(p.TopP),
	)
	return err
}
