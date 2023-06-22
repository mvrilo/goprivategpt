package goprivategpt

import (
	"strings"

	gpt4all "github.com/nomic-ai/gpt4all/gpt4all-bindings/golang"
)

type PrivateGPT struct {
	Model   string
	Threads int
	Tokens  int
	TopK    int
	TopP    float64

	client *gpt4all.Model
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
	l, err := gpt4all.New(p.Model, gpt4all.SetThreads(p.Threads))
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
		gpt4all.SetTokens(p.Tokens),
		gpt4all.SetTopK(p.TopK),
		gpt4all.SetTopP(p.TopP),
	)
	return err
}
